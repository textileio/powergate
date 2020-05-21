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

// New creates a new deal module
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

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs
func (m *Module) Store(ctx context.Context, waddr string, data io.Reader, dcfgs []StorageDealConfig, dur uint64, isCAR bool) (cid.Cid, []StoreResult, error) {
	f, err := ioutil.TempFile(m.cfg.ImportPath, "import-*")
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("error when creating tmpfile: %s", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("closing storing file: %s", err)
		}
	}()
	var size int64
	if size, err = io.Copy(f, data); err != nil {
		return cid.Undef, nil, fmt.Errorf("error when copying data to tmpfile: %s", err)
	}
	gbSize := float64(size) / float64(1<<30)
	ref := api.FileRef{
		Path:  f.Name(),
		IsCAR: isCAR,
	}
	dataCid, err := m.api.ClientImport(ctx, ref)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("error when importing data: %s", err)
	}
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return cid.Undef, nil, err
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
			MinBlocksDuration: dur,
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
	return dataCid, res, nil
}

// Retrieve fetches the data stored in filecoin at a particular cid
func (m *Module) Retrieve(ctx context.Context, waddr string, cid cid.Cid, exportCAR bool) (io.ReadCloser, error) {
	rf, err := ioutil.TempDir(m.cfg.ImportPath, "retrieve-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir for retrieval: %s", err)
	}
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return nil, err
	}
	offers, err := m.api.ClientFindData(ctx, cid)
	if err != nil {
		return nil, err
	}
	if len(offers) == 0 {
		return nil, ErrRetrievalNoAvailableProviders
	}
	fpath := filepath.Join(rf, "ret")
	for _, o := range offers {
		log.Debugf("trying to retrieve data from %s", o.Miner)
		ref := api.FileRef{
			Path:  fpath,
			IsCAR: exportCAR,
		}
		if err = m.api.ClientRetrieve(ctx, o.Order(addr), ref); err != nil {
			log.Infof("retrieving cid %s from %s: %s", cid, o.Miner, err)
			continue
		}
		f, err := os.Open(fpath)
		if err != nil {
			return nil, fmt.Errorf("opening retrieved file: %s", err)
		}
		return f, nil
	}
	return nil, fmt.Errorf("couldn't retrieve data from any miners, last miner err: %s", err)
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
