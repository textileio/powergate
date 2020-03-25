package deals

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/lotus-client/api"
	str "github.com/textileio/lotus-client/chain/store"
)

const (
	initialWait      = time.Second * 5
	chanWriteTimeout = time.Second
)

var (
	ErrRetrievalNoAvailableProviders = errors.New("no providers to retrieve the data")

	log = logging.Logger("deals")
)

// Module exposes storage and monitoring from the market.
type Module struct {
	api API
	cfg *Config
}

// API interacts with a Filecoin full-node
type API interface {
	ClientStartDeal(ctx context.Context, data cid.Cid, addr address.Address, miner address.Address, price types.BigInt, blocksDuration uint64) (*cid.Cid, error)
	ClientImport(ctx context.Context, path string) (cid.Cid, error)
	ClientGetDealInfo(context.Context, cid.Cid) (*api.DealInfo, error)
	ChainNotify(ctx context.Context) (<-chan []*str.HeadChange, error)
	ClientRetrieve(ctx context.Context, order api.RetrievalOrder, path string) error
	ClientFindData(ctx context.Context, root cid.Cid) ([]api.QueryOffer, error)
	StateMarketStorageDeal(context.Context, abi.DealID, types.TipSetKey) (*api.MarketDeal, error)
}

// New creates a new deal module
func New(api API, opts ...Option) (*Module, error) {
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
func (m *Module) Store(ctx context.Context, waddr string, data io.Reader, dcfgs []StorageDealConfig, dur uint64) (cid.Cid, []StoreResult, error) {
	f, err := ioutil.TempFile(m.cfg.ImportPath, "import-*")
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("error when creating tmpfile: %s", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return cid.Undef, nil, fmt.Errorf("error when copying data to tmpfile: %s", err)
	}
	dataCid, err := m.api.ClientImport(ctx, f.Name())
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
		p, err := m.api.ClientStartDeal(ctx, dataCid, addr, maddr, types.NewInt(c.EpochPrice), dur)
		if err != nil {
			log.Errorf("starting deal with %v: %s", c, err)
			res[i] = StoreResult{
				Config: c,
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
func (m *Module) Retrieve(ctx context.Context, waddr string, cid cid.Cid) (io.ReadCloser, error) {
	f, err := ioutil.TempFile(m.cfg.ImportPath, "retrieve-*")
	if err != nil {
		return nil, fmt.Errorf("error when creating tmpfile: %s", err)
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
	for _, o := range offers {
		log.Debugf("trying to retrieve data from %s", o.Miner)
		if err = m.api.ClientRetrieve(ctx, o.Order(addr), f.Name()); err != nil {
			log.Infof("error retrieving cid %s from %s: %s", cid, o.Miner, err)
			continue
		}
		return f, nil
	}
	return nil, fmt.Errorf("couldn't retrieve data from any miners, last miner err: %s", err)
}

// Watch returns a channel with state changes of indicated proposals
func (m *Module) Watch(ctx context.Context, proposals []cid.Cid) (<-chan DealInfo, error) {
	if len(proposals) == 0 {
		return nil, fmt.Errorf("proposals list can't be empty")
	}
	ch := make(chan DealInfo)
	w, err := m.api.ChainNotify(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when listening to chain changes: %s", err)
	}
	go func() {
		defer close(ch)
		currentState := make(map[cid.Cid]*api.DealInfo)
		for {
			select {
			case <-ctx.Done():
				return
			case <-w:
				if err := pushNewChanges(ctx, m.api, currentState, proposals, ch); err != nil {
					log.Errorf("error when pushing new proposal states: %s", err)
				}
			}
		}
	}()
	return ch, nil
}

func pushNewChanges(ctx context.Context, client API, currState map[cid.Cid]*api.DealInfo, proposals []cid.Cid, ch chan<- DealInfo) error {
	for _, pcid := range proposals {
		dinfo, err := client.ClientGetDealInfo(ctx, pcid)
		if err != nil {
			log.Errorf("error when getting deal proposal info %s: %s", pcid, err)
			continue
		}
		if currState[pcid] == nil || (*currState[pcid]).State != dinfo.State {
			currState[pcid] = dinfo
			newState := DealInfo{
				ProposalCid:   dinfo.ProposalCid,
				StateID:       dinfo.State,
				StateName:     storagemarket.DealStates[dinfo.State],
				Miner:         dinfo.Provider.String(),
				PieceRef:      dinfo.PieceRef,
				Size:          dinfo.Size,
				PricePerEpoch: dinfo.PricePerEpoch.Uint64(),
				Duration:      dinfo.Duration,
				DealID:        dinfo.DealID,
			}
			if dinfo.State == storagemarket.StorageDealActive {
				ocd, err := client.StateMarketStorageDeal(ctx, dinfo.DealID, types.EmptyTSK)
				if err != nil {
					log.Errorf("getting on-chain deal info: %s", err)
					continue
				}
				newState.ActivationEpoch = ocd.ActivationEpoch
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
