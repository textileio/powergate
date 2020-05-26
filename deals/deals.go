package deals

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/util"
)

const (
	chanWriteTimeout = time.Second
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
	api *apistruct.FullNodeStruct
	cfg *Config
}

// New creates a new Module.
func New(api *apistruct.FullNodeStruct, opts ...Option) (*Module, error) {
	var cfg Config
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.ImportPath == "" {
		return nil, fmt.Errorf("import path can't be empty")
	}
	return &Module{
		api: api,
		cfg: &cfg,
	}, nil
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
	dataCid, err := m.api.ClientImport(ctx, ref)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("error when importing data: %s", err)
	}
	return dataCid, size, nil
}

// Store create Deal Proposals with all miners indicated in dcfgs. The epoch price
// is automatically calculated considering each miner epoch price and data size.
// The data of dataCid should be already imported to the Filecoin Client or should be
// accessible to it. (e.g: is integrated with an IPFS node).
func (m *Module) Store(ctx context.Context, waddr string, dataCid cid.Cid, size uint64, dcfgs []StorageDealConfig, duration uint64) ([]StoreResult, error) {
	gbSize := float64(size) / float64(1<<30)
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return nil, err
	}
	res := make([]StoreResult, len(dcfgs))
	for i, c := range dcfgs {
		maddr, err := address.NewFromString(c.Miner)
		if err != nil {
			log.Errorf("invalid miner address %v: %s", c, err)
			res[i] = StoreResult{
				Config: c,
			}
			continue
		}
		params := &api.StartDealParams{
			Data: &storagemarket.DataRef{
				Root: dataCid,
			},
			MinBlocksDuration: duration,
			EpochPrice:        types.NewInt(uint64(math.Ceil(2 * gbSize * float64(c.EpochPrice)))),
			Miner:             maddr,
			Wallet:            addr,
		}
		p, err := m.api.ClientStartDeal(ctx, params)
		if err != nil {
			log.Errorf("starting deal with %v: %s", c, err)
			res[i] = StoreResult{
				Config:  c,
				Message: err.Error(),
			}
			continue
		}
		res[i] = StoreResult{
			Config:      c,
			ProposalCid: *p,
			Success:     true,
		}
	}
	return res, nil
}

// Fetch fetches deal data to the underlying blockstore of the Filecoin client.
// This API is meant for clients that use external implementations of blockstores with
// their own API, e.g: IPFS.
func (m *Module) Fetch(ctx context.Context, waddr string, cid cid.Cid) error {
	return m.retrieve(ctx, waddr, cid, nil)
}

// Retrieve retrieves Deal data.
func (m *Module) Retrieve(ctx context.Context, waddr string, cid cid.Cid, CAREncoding bool) (io.ReadCloser, error) {
	rf, err := ioutil.TempDir(m.cfg.ImportPath, "retrieve-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir for retrieval: %s", err)
	}
	ref := api.FileRef{
		Path:  filepath.Join(rf, "ret"),
		IsCAR: CAREncoding,
	}

	if err := m.retrieve(ctx, waddr, cid, &ref); err != nil {
		return nil, fmt.Errorf("retrieving from lotus: %s", err)
	}

	f, err := os.Open(ref.Path)
	if err != nil {
		return nil, fmt.Errorf("opening retrieved file: %s", err)
	}
	return &autodeleteFile{File: f}, nil
}

func (m *Module) retrieve(ctx context.Context, waddr string, cid cid.Cid, ref *api.FileRef) error {
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return err
	}
	offers, err := m.api.ClientFindData(ctx, cid)
	if err != nil {
		return err
	}
	if len(offers) == 0 {
		return ErrRetrievalNoAvailableProviders
	}

	for _, o := range offers {
		if err = m.api.ClientRetrieve(ctx, o.Order(addr), ref); err != nil {
			log.Infof("retrieving cid %s from %s: %s", cid, o.Miner, err)
			continue
		}
		return nil
	}
	return fmt.Errorf("couldn't retrieve data from any miners, last miner err: %s", err)
}

// GetDealStatus returns the current status of the deal, and a flag indicating if the miner of the deal was slashed.
// If the deal doesn't exist, *or has expired* it will return ErrDealNotFound. There's not actual way of distinguishing
// both scenarios in Lotus.
func (m *Module) GetDealStatus(ctx context.Context, pcid cid.Cid) (storagemarket.StorageDealStatus, bool, error) {
	di, err := m.api.ClientGetDealInfo(ctx, pcid)
	if err != nil {
		if strings.Contains(err.Error(), "datastore: key not found") {
			return storagemarket.StorageDealUnknown, false, ErrDealNotFound
		}
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("getting deal info: %s", err)
	}
	md, err := m.api.StateMarketStorageDeal(ctx, di.DealID, types.EmptyTSK)
	if err != nil {
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("get storage state: %s", err)
	}
	return di.State, md.State.SlashEpoch != -1, nil
}

// Watch returns a channel with state changes of indicated proposals
func (m *Module) Watch(ctx context.Context, proposals []cid.Cid) (<-chan DealInfo, error) {
	if len(proposals) == 0 {
		return nil, fmt.Errorf("proposals list can't be empty")
	}
	ch := make(chan DealInfo)
	go func() {
		defer close(ch)
		currentState := make(map[cid.Cid]*api.DealInfo)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(util.AvgBlockTime):
				if err := notifyChanges(ctx, m.api, currentState, proposals, ch); err != nil {
					log.Errorf("pushing new proposal states: %s", err)
				}
			}
		}
	}()
	return ch, nil
}

func notifyChanges(ctx context.Context, client *apistruct.FullNodeStruct, currState map[cid.Cid]*api.DealInfo, proposals []cid.Cid, ch chan<- DealInfo) error {
	for _, pcid := range proposals {
		dinfo, err := client.ClientGetDealInfo(ctx, pcid)
		if err != nil {
			log.Errorf("getting deal proposal info %s: %s", pcid, err)
			continue
		}
		if currState[pcid] == nil || (*currState[pcid]).State != dinfo.State {
			currState[pcid] = dinfo
			newState := DealInfo{
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
					log.Errorf("getting on-chain deal info: %s", err)
					continue
				}
				newState.ActivationEpoch = int64(ocd.State.SectorStartEpoch)
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
