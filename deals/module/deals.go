package module

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/util"
)

const (
	chanWriteTimeout = time.Second

	defaultDealStartOffset = 48 * 60 * 60 / util.EpochDurationSeconds // 48hs
)

var (
	// ErrRetrievalNoAvailableProviders indicates that the data isn't available on any provided
	// to be retrieved.
	ErrRetrievalNoAvailableProviders = errors.New("no providers to retrieve the data")
	// ErrDealNotFound indicates a particular ProposalCid from a deal isn't found on-chain. Currently,
	// in Lotus this indicates that it may never existed on-chain, or it existed but it already expired
	// (currEpoch > StartEpoch+Duration).
	ErrDealNotFound = errors.New("deal not found on-chain")

	log = logging.Logger("deals")
)

// Module exposes storage and monitoring from the market.
type Module struct {
	clientBuilder       lotus.ClientBuilder
	cfg                 *deals.Config
	store               *store
	pollDuration        time.Duration
	dealFinalityTimeout time.Duration
}

// New creates a new Module.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, pollDuration time.Duration, dealFinalityTimeout time.Duration, opts ...deals.Option) (*Module, error) {
	var cfg deals.Config
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}
	m := &Module{
		clientBuilder:       clientBuilder,
		cfg:                 &cfg,
		store:               newStore(ds),
		pollDuration:        pollDuration,
		dealFinalityTimeout: dealFinalityTimeout,
	}
	m.initPendingDeals()
	return m, nil
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

// Store create Deal Proposals with all miners indicated in dcfgs. The epoch price
// is automatically calculated considering each miner epoch price and piece size.
// The data of dataCid should be already imported to the Filecoin Client or should be
// accessible to it. (e.g: is integrated with an IPFS node).
func (m *Module) Store(ctx context.Context, waddr string, dataCid cid.Cid, pieceSize abi.PaddedPieceSize, pieceCid cid.Cid, dcfgs []deals.StorageDealConfig, minDuration uint64) ([]deals.StoreResult, error) {
	if minDuration < util.MinDealDuration {
		return nil, fmt.Errorf("duration %d should be greater or equal to %d", minDuration, util.MinDealDuration)
	}
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return nil, fmt.Errorf("parsing wallet address: %s", err)
	}
	lapi, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	ts, err := lapi.ChainHead(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting chaing head: %s", err)
	}
	res := make([]deals.StoreResult, len(dcfgs))
	for i, c := range dcfgs {
		maddr, err := address.NewFromString(c.Miner)
		if err != nil {
			log.Errorf("invalid miner address %v: %s", c, err)
			res[i] = deals.StoreResult{
				Config: c,
			}
			continue
		}
		dealStartOffset := c.DealStartOffset
		if dealStartOffset == 0 {
			dealStartOffset = defaultDealStartOffset
		}
		params := &api.StartDealParams{
			Data: &storagemarket.DataRef{
				TransferType: storagemarket.TTGraphsync,
				Root:         dataCid,
				PieceCid:     &pieceCid,
				PieceSize:    pieceSize.Unpadded(),
			},
			MinBlocksDuration: minDuration,
			EpochPrice:        big.Div(big.Mul(big.NewIntUnsigned(c.EpochPrice), big.NewIntUnsigned(uint64(pieceSize))), abi.NewTokenAmount(1<<30)),
			Miner:             maddr,
			Wallet:            addr,
			FastRetrieval:     c.FastRetrieval,
			DealStartEpoch:    ts.Height() + abi.ChainEpoch(dealStartOffset),
		}
		p, err := lapi.ClientStartDeal(ctx, params)
		if err != nil {
			log.Errorf("starting deal with %v: %s", c, err)
			res[i] = deals.StoreResult{
				Config:  c,
				Message: err.Error(),
			}
			continue
		}
		res[i] = deals.StoreResult{
			Config:      c,
			ProposalCid: *p,
			Success:     true,
		}
		m.recordDeal(params, *p)
	}
	return res, nil
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

// Watch returns a channel with state changes of indicated proposals.
func (m *Module) Watch(ctx context.Context, proposals []cid.Cid) (<-chan deals.StorageDealInfo, error) {
	if len(proposals) == 0 {
		return nil, fmt.Errorf("proposals list can't be empty")
	}
	ch := make(chan deals.StorageDealInfo)
	go func() {
		defer close(ch)

		currentState := make(map[cid.Cid]*api.DealInfo)

		makeClientAndNotify := func() error {
			client, cls, err := m.clientBuilder(ctx)
			if err != nil {
				return fmt.Errorf("creating lotus client: %s", err)
			}
			if err := notifyChanges(ctx, client, currentState, proposals, ch); err != nil {
				return fmt.Errorf("pushing new proposal states: %s", err)
			}
			cls()
			return nil
		}

		// Notify once so that subscribers get a result quickly
		if err := makeClientAndNotify(); err != nil {
			log.Errorf("creating lotus client and notifying: %s", err)
			return
		}

		// Then notify every m.pollDuration
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.pollDuration):
				if err := makeClientAndNotify(); err != nil {
					log.Errorf("creating lotus client and notifying: %s", err)
					return
				}
			}
		}
	}()
	return ch, nil
}

func notifyChanges(ctx context.Context, lapi *apistruct.FullNodeStruct, currState map[cid.Cid]*api.DealInfo, proposals []cid.Cid, ch chan<- deals.StorageDealInfo) error {
	for _, pcid := range proposals {
		dinfo, err := robustClientGetDealInfo(ctx, lapi, pcid)
		if err != nil {
			return fmt.Errorf("getting deal proposal info %s: %s", pcid, err)
		}
		if currState[pcid] == nil || (*currState[pcid]).State != dinfo.State {
			currState[pcid] = dinfo
			newState, err := fromLotusDealInfo(ctx, lapi, dinfo)
			if err != nil {
				return fmt.Errorf("converting proposal cid %s from lotus deal info: %v", util.CidToString(pcid), err)
			}
			select {
			case <-ctx.Done():
				return nil
			case ch <- newState:
			case <-time.After(chanWriteTimeout):
				log.Warnf("dropping new state since chan is blocked")
			}
		}
	}
	return nil
}

func robustClientGetDealInfo(ctx context.Context, lapi *apistruct.FullNodeStruct, propCid cid.Cid) (*api.DealInfo, error) {
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

func fromLotusDealInfo(ctx context.Context, client *apistruct.FullNodeStruct, dinfo *api.DealInfo) (deals.StorageDealInfo, error) {
	di := deals.StorageDealInfo{
		ProposalCid:   dinfo.ProposalCid,
		StateID:       dinfo.State,
		StateName:     storagemarket.DealStates[dinfo.State],
		Miner:         dinfo.Provider.String(),
		PieceCID:      dinfo.PieceCID,
		Size:          dinfo.Size,
		PricePerEpoch: dinfo.PricePerEpoch.Uint64(),
		Duration:      dinfo.Duration,
		DealID:        uint64(dinfo.DealID),
		Message:       dinfo.Message,
	}
	if dinfo.State == storagemarket.StorageDealActive {
		ocd, err := client.StateMarketStorageDeal(ctx, dinfo.DealID, types.EmptyTSK)
		if err != nil {
			return deals.StorageDealInfo{}, fmt.Errorf("getting on-chain deal info: %s", err)
		}
		di.ActivationEpoch = int64(ocd.State.SectorStartEpoch)
		di.StartEpoch = uint64(ocd.Proposal.StartEpoch)
		di.Duration = uint64(ocd.Proposal.EndEpoch) - uint64(ocd.Proposal.StartEpoch) + 1
	}
	return di, nil
}

type autodeleteFile struct {
	*os.File
}

func (af *autodeleteFile) Close() error {
	if err := af.File.Close(); err != nil {
		return fmt.Errorf("closing retrieval file: %s", err)
	}
	if err := os.Remove(af.File.Name()); err != nil {
		return fmt.Errorf("autodeleting retrieval file: %s", err)
	}
	return nil
}
