package module

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/deals/module/dealwatcher"
	"github.com/textileio/powergate/v2/deals/module/store"
	"github.com/textileio/powergate/v2/lotus"
	"go.opentelemetry.io/otel/metric"
)

var (
	log = logging.Logger("deals")
)

// Module exposes storage and monitoring from the market.
type Module struct {
	clientBuilder       lotus.ClientBuilder
	cfg                 *deals.Config
	store               *store.Store
	dealWatcher         *dealwatcher.DealWatcher
	pollDuration        time.Duration
	dealFinalityTimeout time.Duration

	metricDealTracking      metric.Int64UpDownCounter
	metricRetrievalTracking metric.Int64UpDownCounter
}

// New creates a new Module.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, pollDuration time.Duration, dealFinalityTimeout time.Duration, opts ...deals.Option) (*Module, error) {
	var cfg deals.Config
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}

	log.Infof("creating deal watcher")
	dw, err := dealwatcher.New(clientBuilder)
	if err != nil {
		return nil, fmt.Errorf("creating deal watcher: %s", err)
	}
	m := &Module{
		clientBuilder:       clientBuilder,
		cfg:                 &cfg,
		store:               store.New(ds),
		pollDuration:        pollDuration,
		dealFinalityTimeout: dealFinalityTimeout,
		dealWatcher:         dw,
	}
	m.initMetrics()

	log.Infof("resuming pending records")
	if err := m.resumeWatchingPendingRecords(); err != nil {
		return nil, fmt.Errorf("resuming watching pending records: %s", err)
	}

	return m, nil
}

// GetDealInfo returns info about a deal. If the deal isn't active on-chain,
// it returns ErrDealNotFound.
func (m *Module) GetDealInfo(ctx context.Context, dealID uint64) (api.MarketDeal, error) {
	lapi, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return api.MarketDeal{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	id := abi.DealID(dealID)
	smd, err := lapi.StateMarketStorageDeal(ctx, id, types.EmptyTSK)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return api.MarketDeal{}, ErrDealNotFound
		}
		return api.MarketDeal{}, fmt.Errorf("getting deal info: %s", err)
	}
	if smd == nil {
		return api.MarketDeal{}, fmt.Errorf("deal on-chain information is empty")
	}

	return *smd, nil
}

// CalculateDealPiece calculates the size and CommP for a data cid.
func (m *Module) CalculateDealPiece(ctx context.Context, c cid.Cid) (api.DataCIDSize, error) {
	lapi, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return api.DataCIDSize{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	dsz, err := lapi.ClientDealPieceCID(ctx, c)
	if err != nil {
		return api.DataCIDSize{}, fmt.Errorf("calculating data size: %s", err)
	}
	return dsz, nil
}

// Import imports raw data in the Filecoin client. The isCAR flag indicates if the data
// is already in CAR format, so it shouldn't be encoded into a UnixFS DAG in the Filecoin client.
// It returns the imported data cid and the data size.
func (m *Module) Import(ctx context.Context, data io.Reader, isCAR bool) (cid.Cid, int64, error) {
	f, err := ioutil.TempFile(m.cfg.ImportPath, "import-*")
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("error when creating tmpfile: %s", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("closing storing file: %s", err)
		}
	}()
	var size int64
	if size, err = io.Copy(f, data); err != nil {
		return cid.Undef, 0, fmt.Errorf("error when copying data to tmpfile: %s", err)
	}
	ref := api.FileRef{
		Path:  f.Name(),
		IsCAR: isCAR,
	}
	api, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	res, err := api.ClientImport(ctx, ref)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("error when importing data: %s", err)
	}
	return res.Root, size, nil
}

// Close gracefully shutdowns the deals module.
func (m *Module) Close() error {
	if err := m.dealWatcher.Close(); err != nil {
		return fmt.Errorf("closing deal watcher: %s", err)
	}
	return nil
}

func robustClientGetDealInfo(ctx context.Context, lapi *api.FullNodeStruct, propCid cid.Cid) (*api.DealInfo, error) {
	di, err := lapi.ClientGetDealInfo(ctx, propCid)
	if err != nil {
		return nil, fmt.Errorf("client get deal info: %s", err)
	}

	// Workaround ClientGetDealInfo giving unreliable status.
	// If the state isn't Active, double-check with on-chain API.
	if di.DealID != 0 && di.State != storagemarket.StorageDealActive {
		smd, err := lapi.StateMarketStorageDeal(ctx, di.DealID, types.EmptyTSK)
		if err != nil {
			// If the DealID is not found, most probably is because the Lotus
			// node isn't yet synced. Since all this logic is a workaround
			// for a problem, make an exception and just return what
			// ClientGetDealInfo said.
			if strings.Contains(err.Error(), "not found") {
				return di, nil
			}
			return nil, fmt.Errorf("state market storage deal: %s", err)
		}
		if smd.State.SectorStartEpoch > 0 {
			di.State = storagemarket.StorageDealActive
		}
	}
	return di, nil
}
