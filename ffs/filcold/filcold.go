package filcold

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logger "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-car"
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
	l     ffs.CidLogger
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// FilChain is an abstraction of a Filecoin node to get information of the network.
type FilChain interface {
	GetHeight(context.Context) (uint64, error)
}

// New returns a new FilCold instance
func New(ms ffs.MinerSelector, dm *deals.Module, dag format.DAGService, chain FilChain, l ffs.CidLogger) *FilCold {
	return &FilCold{
		ms:    ms,
		dm:    dm,
		dag:   dag,
		chain: chain,
		l:     l,
	}
}

// Retrieve returns the original data Cid, from the CAR encoded data Cid. The returned Cid is available in the
// car.Store received as a parameter.
func (fc *FilCold) Retrieve(ctx context.Context, dataCid cid.Cid, cs car.Store, waddr string) (cid.Cid, error) {
	carR, err := fc.dm.Retrieve(ctx, waddr, dataCid, true)
	if err != nil {
		return cid.Undef, fmt.Errorf("retrieving from deal module: %s", err)
	}
	defer func() {
		if err := carR.Close(); err != nil {
			log.Errorf("closing reader from deal retrieve: %s", err)
		}
	}()
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
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, cfg ffs.FilConfig) (ffs.FilInfo, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: cfg.ExcludedMiners,
		CountryCodes:   cfg.CountryCodes,
		TrustedMiners:  cfg.TrustedMiners,
	}
	cfgs, err := makeDealConfigs(ctx, fc.ms, cfg.RepFactor, f)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("making deal configs: %s", err)
	}
	props, err := fc.makeDeals(ctx, c, cfgs, cfg)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("executing deals: %s", err)
	}

	return ffs.FilInfo{
		DataCid:   c,
		Proposals: props,
	}, nil
}

// IsFilDealActive returns true if a deal is considered active on-chain, false otherwise.
func (fc *FilCold) IsFilDealActive(ctx context.Context, proposalCid cid.Cid) (bool, error) {
	status, slashed, err := fc.dm.GetDealStatus(ctx, proposalCid)
	if err == deals.ErrDealNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("getting deal state for %s: %s", proposalCid, err)
	}
	return !slashed && status == storagemarket.StorageDealActive, nil

}

// EnsureRenewals analyzes a FilInfo state for a Cid and executes renewals considering the FilConfig desired configuration.
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, cfg ffs.FilConfig) (ffs.FilInfo, error) {
	var activeMiners []string
	for _, p := range inf.Proposals {
		activeMiners = append(activeMiners, p.Miner)
	}
	height, err := fc.chain.GetHeight(ctx)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("get current filecoin height: %s", err)
	}
	var renewable []ffs.FilStorage
	for _, p := range inf.Proposals {
		if p.Renewed {
			continue
		}
		expiry := p.ActivationEpoch + p.Duration
		renewalHeight := expiry - int64(cfg.Renew.Threshold)
		if uint64(renewalHeight) <= height {
			renewable = append(renewable, p)
		}
	}

	// Calculate how many active deals aren't expiring soon.
	youngActiveDeals := len(inf.Proposals) - len(renewable)
	// Calculate how many of the renewable (soon to be expired) deals should be renewed
	// so we have the desired RepFactor.
	numToBeRenewed := cfg.RepFactor - youngActiveDeals
	if numToBeRenewed <= 0 {
		// Nothing to be renewed to ensure RepFactor,
		// most prob the RepFactor was decreased.
		return inf, nil
	}
	if numToBeRenewed > len(renewable) {
		// We need even more deals than renewable to ensure RepFactor,
		// that's job of the repair module. We renew as many as can be
		// renewed.
		numToBeRenewed = len(renewable)
	}
	toRenew := renewable[:numToBeRenewed]
	for i, p := range toRenew {
		newProposal, err := fc.renewDeal(ctx, c, p, activeMiners, cfg)
		if err != nil {
			log.Errorf("renewing deal %s: %s", p.ProposalCid, err)
			continue
		}
		inf.Proposals = append(inf.Proposals, newProposal)
		inf.Proposals[i].Renewed = true
	}

	return inf, nil
}

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, p ffs.FilStorage, activeMiners []string, fcfg ffs.FilConfig) (ffs.FilStorage, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: activeMiners,
	}
	dealConfig, err := makeDealConfigs(ctx, fc.ms, 1, f)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("making new deal config: %s", err)
	}

	props, err := fc.makeDeals(ctx, c, dealConfig, fcfg)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("executing renewed deal: %s", err)
	}
	if len(props) != 1 {
		return ffs.FilStorage{}, fmt.Errorf("unsuccessful renewal")
	}
	return props[0], nil
}

func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, cfgs []deals.StorageDealConfig, fcfg ffs.FilConfig) ([]ffs.FilStorage, error) {
	r := ipldToFileTransform(ctx, fc.dag, c)

	for _, cfg := range cfgs {
		fc.l.Log(ctx, c, "Proposing deal to miner %s with %d fil per epoch...", cfg.Miner, cfg.EpochPrice)
	}

	var sres []deals.StoreResult
	dataCid, sres, err := fc.dm.Store(ctx, fcfg.Addr, r, cfgs, uint64(fcfg.DealDuration), true)
	if err != nil {
		return nil, fmt.Errorf("storing deals in deal module: %s", err)
	}
	if dataCid != c {
		return nil, fmt.Errorf("stored data cid doesn't match with sent data")
	}

	proposals, err := fc.waitForDeals(ctx, c, sres, fcfg.DealDuration)
	if err != nil {
		return nil, fmt.Errorf("waiting for deals to finish: %s", err)
	}
	return proposals, nil
}

func (fc *FilCold) waitForDeals(ctx context.Context, c cid.Cid, storeResults []deals.StoreResult, duration int64) ([]ffs.FilStorage, error) {
	notDone := make(map[cid.Cid]struct{})
	var inProgressDeals []cid.Cid
	for _, d := range storeResults {
		if !d.Success {
			fc.l.Log(ctx, c, "Proposal with miner %s failed.", d.Config.Miner)
			log.Warnf("failed store result")
			continue
		}
		inProgressDeals = append(inProgressDeals, d.ProposalCid)
		notDone[d.ProposalCid] = struct{}{}
	}
	if len(inProgressDeals) == 0 {
		return nil, fmt.Errorf("all proposed deals where rejected")
	}

	fc.l.Log(ctx, c, "Watching in-progress deals unfold...")
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
			fc.l.Log(ctx, c, "Deal %d with miner %s is active on-chain", di.DealID, di.Miner)
		} else if di.StateID == storagemarket.StorageDealError || di.StateID == storagemarket.StorageDealFailing {
			log.Errorf("deal %d failed with state %s", di.DealID, storagemarket.DealStates[di.StateID])
			delete(activeProposals, di.ProposalCid)
			delete(notDone, di.ProposalCid)
			fc.l.Log(ctx, c, "Deal %d with miner %s failed and won't be active on-chain", di.DealID, di.Miner)
		} else {
			fc.l.Log(ctx, c, "Deal with miner %s changed state to %s", di.Miner, storagemarket.DealStates[di.StateID])
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
	fc.l.Log(ctx, c, "Finished all in-progress deals reached final state.")
	return res, nil
}

func ipldToFileTransform(ctx context.Context, dag format.DAGService, c cid.Cid) io.Reader {
	r, w := io.Pipe()
	go func() {
		if err := car.WriteCar(ctx, dag, []cid.Cid{c}, w); err != nil {
			log.Errorf("writing car file: %s", err)
			if err := w.CloseWithError(err); err != nil {
				log.Errorf("closing with error: %s", err)
			}
		}
		if err := w.Close(); err != nil {
			log.Errorf("closing writer in ipld to file transform: %s", err)
		}
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
