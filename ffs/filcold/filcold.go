package filcold

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logger.Logger("ffs-filcold")
)

// FilCold is a ColdStorage implementation which saves data in the Filecoin network.
// It assumes the underlying Filecoin client has access to an IPFS node where data is stored.
type FilCold struct {
	ms    ffs.MinerSelector
	dm    *deals.Module
	ipfs  iface.CoreAPI
	chain FilChain
	l     ffs.CidLogger
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// FilChain is an abstraction of a Filecoin node to get information of the network.
type FilChain interface {
	GetHeight(context.Context) (uint64, error)
}

// New returns a new FilCold instance.
func New(ms ffs.MinerSelector, dm *deals.Module, ipfs iface.CoreAPI, chain FilChain, l ffs.CidLogger) *FilCold {
	return &FilCold{
		ms:    ms,
		dm:    dm,
		ipfs:  ipfs,
		chain: chain,
		l:     l,
	}
}

// Fetch fetches the stored Cid data.The data will be considered available
// to the underlying blockstore.
func (fc *FilCold) Fetch(ctx context.Context, dataCid cid.Cid, waddr string) error {
	if err := fc.dm.Fetch(ctx, waddr, dataCid); err != nil {
		return fmt.Errorf("fetching from deal module: %s", err)
	}
	return nil
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation. It returns a slice of ProposalCids that were correctly
// started, and a slice of with Proposal Cids rejected. Returned proposed deals can be tracked
// with the WaitForDeals API.
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, cfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, uint64, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: cfg.ExcludedMiners,
		CountryCodes:   cfg.CountryCodes,
		TrustedMiners:  cfg.TrustedMiners,
		MaxPrice:       cfg.MaxPrice,
	}
	cfgs, err := makeDealConfigs(ctx, fc.ms, cfg.RepFactor, f)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("making deal configs: %s", err)
	}
	size, err := fc.getCidSize(ctx, c)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("getting cid cummulative size: %s", err)
	}
	okDeals, failedStartingDeals, err := fc.makeDeals(ctx, c, size, cfgs, cfg)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("starting deals: %s", err)
	}
	if len(okDeals) == 0 {
		return nil, failedStartingDeals, 0, fmt.Errorf("all proposed deals where rejected")
	}
	return okDeals, failedStartingDeals, size, nil
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
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, cfg ffs.FilConfig) (ffs.FilInfo, []ffs.DealError, error) {
	height, err := fc.chain.GetHeight(ctx)
	if err != nil {
		return ffs.FilInfo{}, nil, fmt.Errorf("get current filecoin height: %s", err)
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
		return inf, nil, nil
	}
	if numToBeRenewed > len(renewable) {
		// We need even more deals than renewable to ensure RepFactor,
		// that's job of the repair module. We renew as many as can be
		// renewed.
		numToBeRenewed = len(renewable)
	}

	toRenew := renewable[:numToBeRenewed]
	var retErrors []ffs.DealError
	for i, p := range toRenew {
		newProposal, errors, err := fc.renewDeal(ctx, c, inf.Size, p, cfg)
		if err != nil {
			log.Errorf("renewing deal %s: %s", p.ProposalCid, err)
			continue
		}
		inf.Proposals = append(inf.Proposals, newProposal)
		inf.Proposals[i].Renewed = true
		retErrors = append(retErrors, errors...)
	}

	return inf, retErrors, nil
}

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, size uint64, p ffs.FilStorage, fcfg ffs.FilConfig) (ffs.FilStorage, []ffs.DealError, error) {
	f := ffs.MinerSelectorFilter{
		TrustedMiners: []string{p.Miner},
		MaxPrice:      fcfg.MaxPrice,
	}
	dealConfig, err := makeDealConfigs(ctx, fc.ms, 1, f)
	if err != nil {
		return ffs.FilStorage{}, nil, fmt.Errorf("making new deal config: %s", err)
	}

	okDeals, failedStartedDeals, err := fc.makeDeals(ctx, c, size, dealConfig, fcfg)
	if err != nil {
		return ffs.FilStorage{}, nil, fmt.Errorf("executing renewed deal: %s", err)
	}
	if len(okDeals) != 1 {
		fc.l.Log(ctx, c, "Starting renewal deal proposal failed: %s", failedStartedDeals[0].Message)
		return ffs.FilStorage{}, failedStartedDeals, fmt.Errorf("starting renew deal proposal: %s", err)
	}
	successDeals, failedExecutingDeals, err := fc.WaitForDeals(ctx, c, okDeals)
	if err != nil {
		return ffs.FilStorage{}, failedExecutingDeals, fmt.Errorf("waiting for renweal deal: %s", err)
	}
	if len(successDeals) != 1 {
		fc.l.Log(ctx, c, "Renewal deal execution errored: %s", failedStartedDeals[0].Message)
		return ffs.FilStorage{}, failedExecutingDeals, fmt.Errorf("unsuccessful renewal")
	}
	return successDeals[0], nil, nil
}

// makeDeals starts deals with the specified miners. It returns a slice with all the ProposalCids
// that were started successfully, and a slice of DealError with deals that failed to be started.
func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, size uint64, cfgs []deals.StorageDealConfig, fcfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, error) {
	for _, cfg := range cfgs {
		fc.l.Log(ctx, c, "Proposing deal to miner %s with %d fil per epoch...", cfg.Miner, cfg.EpochPrice)
	}

	sres, err := fc.dm.Store(ctx, fcfg.Addr, c, size, cfgs, uint64(fcfg.DealDuration))
	if err != nil {
		return nil, nil, fmt.Errorf("storing deals in deal module: %s", err)
	}
	var okDeals []cid.Cid
	var failedDeals []ffs.DealError
	for _, r := range sres {
		if !r.Success {
			fc.l.Log(ctx, c, "Proposal with miner %s failed: %s", r.Config.Miner, r.Message)
			log.Warnf("failed store result: %s", r.Message)
			de := ffs.DealError{
				ProposalCid: r.ProposalCid,
				Message:     r.Message,
				Miner:       r.Config.Miner,
			}
			failedDeals = append(failedDeals, de)
			continue
		}
		okDeals = append(okDeals, r.ProposalCid)
	}
	return okDeals, failedDeals, nil
}

func (fc *FilCold) WaitForDeals(ctx context.Context, c cid.Cid, proposals []cid.Cid) ([]ffs.FilStorage, []ffs.DealError, error) {
	var errors []ffs.DealError
	notDone := make(map[cid.Cid]struct{})
	for _, d := range proposals {
		notDone[d] = struct{}{}
	}
	fc.l.Log(ctx, c, "Watching deals unfold...")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDi, err := fc.dm.Watch(ctx, proposals)
	if err != nil {
		return nil, errors, err
	}

	activeProposals := make(map[cid.Cid]*ffs.FilStorage)
	for di := range chDi {
		log.Infof("watching pending %d deals unfold...", len(notDone))
		if _, ok := notDone[di.ProposalCid]; !ok {
			continue
		}
		if di.StateID == storagemarket.StorageDealActive {
			activeProposals[di.ProposalCid] = &ffs.FilStorage{
				ProposalCid:     di.ProposalCid,
				Duration:        int64(di.Duration),
				Miner:           di.Miner,
				ActivationEpoch: di.ActivationEpoch,
				EpochPrice:      di.PricePerEpoch,
			}
			delete(notDone, di.ProposalCid)
			fc.l.Log(ctx, c, "Deal %d with miner %s is active on-chain", di.DealID, di.Miner)
		} else if di.StateID == storagemarket.StorageDealError || di.StateID == storagemarket.StorageDealFailing {
			log.Errorf("deal %d failed with state %s: %s", di.DealID, storagemarket.DealStates[di.StateID], di.Message)
			delete(activeProposals, di.ProposalCid)
			delete(notDone, di.ProposalCid)
			fc.l.Log(ctx, c, "DealID %d with miner %s failed and won't be active on-chain: %s", di.DealID, di.Miner, di.Message)
			errors = append(errors, ffs.DealError{ProposalCid: di.ProposalCid, Miner: di.Miner, Message: di.Message})
		} else {
			fc.l.Log(ctx, c, "Deal with miner %s changed state to %s", di.Miner, storagemarket.DealStates[di.StateID])
		}
		if len(notDone) == 0 {
			break
		}
	}

	if len(activeProposals) == 0 {
		return nil, errors, fmt.Errorf("all accepted proposals failed before becoming active")
	}
	res := make([]ffs.FilStorage, 0, len(activeProposals))
	for _, v := range activeProposals {
		res = append(res, *v)
	}
	fc.l.Log(ctx, c, "All deals reached final state.")
	return res, errors, nil
}

func (fc *FilCold) getCidSize(ctx context.Context, c cid.Cid) (uint64, error) {
	s, err := fc.ipfs.Object().Stat(ctx, path.IpfsPath(c))
	if err != nil {
		return 0, fmt.Errorf("calling ipfs object stat: %s", err)
	}
	return uint64(s.CumulativeSize), nil
}

func makeDealConfigs(ctx context.Context, ms ffs.MinerSelector, cntMiners int, f ffs.MinerSelectorFilter) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetMiners(cntMiners, f)
	if err != nil {
		return nil, fmt.Errorf("getting miners from minerselector: %s", err)
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice}
	}
	return res, nil
}
