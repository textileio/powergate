package filcold

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logger.Logger("ffs-filcold")
)

// FilCold is a ffs.ColdStorage implementation that stores data in Filecoin.
type FilCold struct {
	ms    ffs.MinerSelector
	dm    *deals.Module
	dag   format.DAGService
	chain FilChain
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// FilChain is an abstraction of a Filecoin node to get information of the network.
type FilChain interface {
	GetHeight(context.Context) (uint64, error)
}

// New returns a new FilCold instance
func New(ms ffs.MinerSelector, dm *deals.Module, dag format.DAGService, chain FilChain) *FilCold {
	return &FilCold{
		ms:    ms,
		dm:    dm,
		dag:   dag,
		chain: chain,
	}
}

// Retrieve returns the original data Cid, from the CAR encoded data Cid. The returned Cid is available in the
// car.Store received as a parameter.
func (fc *FilCold) Retrieve(ctx context.Context, dataCid cid.Cid, cs car.Store, waddr string) (cid.Cid, error) {
	carR, err := fc.dm.Retrieve(ctx, waddr, dataCid, true)
	if err != nil {
		return cid.Undef, fmt.Errorf("retrieving from deal module: %s", err)
	}
	defer carR.Close()
	h, err := car.LoadCar(cs, carR)
	if err != nil {
		return cid.Undef, fmt.Errorf("loading car to carstore: %s", err)
	}
	if len(h.Roots) != 1 {
		return cid.Undef, fmt.Errorf("car header doesn't have a single root: %d", len(h.Roots))
	}
	return h.Roots[0], nil
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation.
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, waddr string, cfg ffs.FilConfig) (ffs.FilInfo, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: cfg.ExcludedMiners,
		CountryCodes:   cfg.CountryCodes,
	}
	cfgs, err := makeDealConfigs(ctx, fc.ms, cfg.RepFactor, f)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("making deal configs: %s", err)
	}
	props, err := fc.makeDeals(ctx, c, cfgs, waddr, cfg)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("executing deals: %s", err)
	}

	return ffs.FilInfo{
		DataCid:   c,
		Proposals: props,
	}, nil
}

func (fc *FilCold) IsFilDealActive(ctx context.Context, proposalCid cid.Cid) (bool, error) {
	status, slashed, err := fc.dm.GetDealStatus(ctx, proposalCid)
	if err != nil {
		return false, fmt.Errorf("getting deal state for %s: %s", proposalCid, err)
	}
	return !slashed && status == storagemarket.StorageDealActive, nil

}

// EnsureRenewals analyzes a FilInfo state for a Cid and executes renewals considering the FilConfig desired configuration.
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, waddr string, fcfg ffs.FilConfig) (ffs.FilInfo, error) {
	var activeMiners []string
	for _, p := range inf.Proposals {
		activeMiners = append(activeMiners, p.Miner)
	}
	height, err := fc.chain.GetHeight(ctx)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("get current filecoin height: %s", err)
	}
	for i, p := range inf.Proposals {
		if p.Renewed {
			continue
		}
		expiry := p.ActivationEpoch + p.Duration
		renewalHeight := expiry - int64(fcfg.Renew.Threshold)
		if uint64(renewalHeight) <= height {
			newProposal, err := fc.renewDeal(ctx, c, waddr, p, activeMiners, fcfg)
			if err != nil {
				log.Errorf("renewing deal %s: %s", p.ProposalCid, err)
				continue
			}
			inf.Proposals = append(inf.Proposals, newProposal)
			inf.Proposals[i].Renewed = true
		}
	}
	return inf, nil
}

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, waddr string, p ffs.FilStorage, activeMiners []string, fcfg ffs.FilConfig) (ffs.FilStorage, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: activeMiners,
	}
	dealConfig, err := makeDealConfigs(ctx, fc.ms, 1, f)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("making new deal config: %s", err)
	}

	props, err := fc.makeDeals(ctx, c, dealConfig, waddr, fcfg)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("executing renewed deal: %s", err)
	}
	if len(props) != 1 {
		return ffs.FilStorage{}, fmt.Errorf("unsuccessful renewal")
	}
	return props[0], nil
}

func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, cfgs []deals.StorageDealConfig, waddr string, fcfg ffs.FilConfig) ([]ffs.FilStorage, error) {
	r := ipldToFileTransform(ctx, fc.dag, c)

	var sres []deals.StoreResult
	dataCid, sres, err := fc.dm.Store(ctx, waddr, r, cfgs, uint64(fcfg.DealDuration), true)
	if err != nil {
		return nil, fmt.Errorf("storing deals in deal module: %s", err)
	}
	if dataCid != c {
		return nil, fmt.Errorf("stored data cid doesn't match with sent data")
	}

	proposals, err := fc.waitForDeals(ctx, sres, fcfg.DealDuration)
	if err != nil {
		return nil, fmt.Errorf("waiting for deals to finish: %s", err)
	}
	return proposals, nil
}

func (fc *FilCold) waitForDeals(ctx context.Context, storeResults []deals.StoreResult, duration int64) ([]ffs.FilStorage, error) {
	notDone := make(map[cid.Cid]struct{})
	var inProgressDeals []cid.Cid
	for _, d := range storeResults {
		if !d.Success {
			log.Warnf("failed store result")
			continue
		}
		inProgressDeals = append(inProgressDeals, d.ProposalCid)
		notDone[d.ProposalCid] = struct{}{}
	}
	if len(inProgressDeals) == 0 {
		return nil, fmt.Errorf("all proposed deals where rejected")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDi, err := fc.dm.Watch(ctx, inProgressDeals)
	if err != nil {
		return nil, err
	}

	activeProposals := make(map[cid.Cid]*ffs.FilStorage)
	for di := range chDi {
		log.Infof("watching pending %d deals unfold...", len(notDone))
		if _, ok := notDone[di.ProposalCid]; !ok {
			continue
		}
		if di.StateID == storagemarket.StorageDealActive {
			var d *deals.StoreResult
			for _, dr := range storeResults {
				if dr.ProposalCid == di.ProposalCid {
					d = &dr
					break
				}
			}
			if d == nil {
				return nil, fmt.Errorf("deal watcher return unasked proposal, this must never happen")
			}
			activeProposals[di.ProposalCid] = &ffs.FilStorage{
				ProposalCid:     di.ProposalCid,
				Duration:        duration,
				Miner:           d.Config.Miner,
				ActivationEpoch: di.ActivationEpoch,
			}
			delete(notDone, di.ProposalCid)
		} else if di.StateID == storagemarket.StorageDealError || di.StateID == storagemarket.StorageDealFailing {
			log.Errorf("deal %s failed with state %s", di.ProposalCid, storagemarket.DealStates[di.StateID])
			delete(activeProposals, di.ProposalCid)
			delete(notDone, di.ProposalCid)
		}
		if len(notDone) == 0 {
			break
		}
	}

	if len(activeProposals) == 0 {
		return nil, fmt.Errorf("all accepted proposals failed before becoming active")
	}
	res := make([]ffs.FilStorage, 0, len(activeProposals))
	for _, v := range activeProposals {
		res = append(res, *v)
	}
	return res, nil
}

func ipldToFileTransform(ctx context.Context, dag format.DAGService, c cid.Cid) io.Reader {
	r, w := io.Pipe()
	go func() {
		if err := car.WriteCar(ctx, dag, []cid.Cid{c}, w); err != nil {
			w.CloseWithError(err)
		}
		w.Close()
	}()
	return r
}

func makeDealConfigs(ctx context.Context, ms ffs.MinerSelector, cantMiners int, f ffs.MinerSelectorFilter) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetMiners(cantMiners, f)
	if err != nil {
		return nil, fmt.Errorf("getting miners from minerselector: %s", err)
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice}
	}
	return res, nil
}
