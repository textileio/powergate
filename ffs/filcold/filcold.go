package filcold

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-padreader"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logger "github.com/ipfs/go-log/v2"
	dag "github.com/ipfs/go-merkledag"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/ipld/go-car"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/deals/module"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logger.Logger("ffs-filcold")
)

// FilCold is a ColdStorage implementation which saves data in the Filecoin network.
// It assumes the underlying Filecoin client has access to an IPFS node where data is stored.
type FilCold struct {
	ms    ffs.MinerSelector
	dm    *dealsModule.Module
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
func New(ms ffs.MinerSelector, dm *dealsModule.Module, ipfs iface.CoreAPI, chain FilChain, l ffs.CidLogger) *FilCold {
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
func (fc *FilCold) Fetch(ctx context.Context, pyCid cid.Cid, piCid *cid.Cid, waddr string, miners []string, maxPrice uint64, selector string) (ffs.FetchInfo, error) {
	miner, events, err := fc.dm.Fetch(ctx, waddr, pyCid, piCid, miners)
	if err != nil {
		return ffs.FetchInfo{}, fmt.Errorf("fetching from deal module: %s", err)
	}
	var fundsSpent uint64
	for e := range events {
		strEvent := retrievalmarket.ClientEvents[e.Event]
		strDealStatus := retrievalmarket.DealStatuses[e.Status]
		fc.l.Log(ctx, "Event: %s, bytes received %d, funds spent: %d attoFil, status: %s ", strEvent, e.BytesReceived, e.FundsSpent, strDealStatus)
		fundsSpent = e.FundsSpent.Uint64()

	}
	return ffs.FetchInfo{RetrievedMiner: miner, FundsSpent: fundsSpent}, nil
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation. It returns a slice of ProposalCids that were correctly
// started, and a slice of with Proposal Cids rejected. Returned proposed deals can be tracked
// with the WaitForDeal API.
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, cfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, uint64, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: cfg.ExcludedMiners,
		CountryCodes:   cfg.CountryCodes,
		TrustedMiners:  cfg.TrustedMiners,
		MaxPrice:       cfg.MaxPrice,
	}
	cfgs, err := makeDealConfigs(fc.ms, cfg.RepFactor, f, cfg.FastRetrieval)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("making deal configs: %s", err)
	}

	fc.l.Log(ctx, "Calculating piece size...")
	size, err := fc.calculatePieceSize(ctx, c)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("getting cid cummulative size: %s", err)
	}
	fc.l.Log(ctx, "Estimated piece size is %d bytes.", size)

	okDeals, failedStartingDeals, err := fc.makeDeals(ctx, c, size, cfgs, cfg)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("starting deals: %s", err)
	}
	return okDeals, failedStartingDeals, size, nil
}

// IsFilDealActive returns true if a deal is considered active on-chain, false otherwise.
func (fc *FilCold) IsFilDealActive(ctx context.Context, proposalCid cid.Cid) (bool, error) {
	status, slashed, err := fc.dm.GetDealStatus(ctx, proposalCid)
	if err == module.ErrDealNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("getting deal state for %s: %s", proposalCid, err)
	}
	return !slashed && status == storagemarket.StorageDealActive, nil
}

// EnsureRenewals analyzes a FilInfo state for a Cid and executes renewals considering the FilConfig desired configuration.
// It returns an updated FilInfo for the Cid. All prevous Proposals in the received FilInfo are kept, only flagging the ones
// that got renewed with Renewed=true. New deals from renewals are added to the returned FilInfo.
// Note: Most probably all this code should change in the future, when Filecoin supports telling the miner which deal is about to
// expire that we're interested in extending the deal duration. Now we should make a new deal from scratch (send data, etc).
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, cfg ffs.FilConfig) (ffs.FilInfo, []ffs.DealError, error) {
	height, err := fc.chain.GetHeight(ctx)
	if err != nil {
		return ffs.FilInfo{}, nil, fmt.Errorf("get current filecoin height: %s", err)
	}

	var renewable []ffs.FilStorage
	for _, p := range inf.Proposals {
		// If this deal was already renewed, we can ignore it will
		// soon expire since we already handled it.
		if p.Renewed {
			continue
		}
		// In imported deals data, we might have missing information
		// about start and/or duration. If that's the case, ignore
		// them.
		if p.StartEpoch == 0 || p.Duration == 0 {
			continue
		}
		expiry := int64(p.StartEpoch) + p.Duration
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

	newInf := ffs.FilInfo{
		DataCid:   inf.DataCid,
		Size:      inf.Size,
		Proposals: make([]ffs.FilStorage, len(inf.Proposals)),
	}
	for i, p := range inf.Proposals {
		newInf.Proposals[i] = p
	}

	// Manually imported doesn't provide the piece size.
	// Re-calculate it if necessary. If present, just re-use that value.
	if inf.Size == 0 {
		size, err := fc.calculatePieceSize(ctx, inf.DataCid)
		if err != nil {
			return ffs.FilInfo{}, nil, fmt.Errorf("can't recalculate piece size: %s", err)
		}
		inf.Size = size
	}

	toRenew := renewable[:numToBeRenewed]
	var newDealErrors []ffs.DealError
	for i, p := range toRenew {
		var dealError ffs.DealError
		newProposal, err := fc.renewDeal(ctx, c, inf.Size, p, cfg)
		if err != nil {
			if errors.As(err, &dealError) {
				newDealErrors = append(newDealErrors, dealError)
				continue
			}
			continue
		}
		newInf.Proposals = append(newInf.Proposals, newProposal)
		newInf.Proposals[i].Renewed = true
	}

	return newInf, newDealErrors, nil
}

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, size uint64, p ffs.FilStorage, fcfg ffs.FilConfig) (ffs.FilStorage, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: fcfg.ExcludedMiners,
		CountryCodes:   fcfg.CountryCodes,
		TrustedMiners:  []string{p.Miner},
		MaxPrice:       fcfg.MaxPrice,
	}
	dealConfig, err := makeDealConfigs(fc.ms, 1, f, fcfg.FastRetrieval)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("making new deal config: %s", err)
	}

	okDeals, failedStartedDeals, err := fc.makeDeals(ctx, c, size, dealConfig, fcfg)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("executing renewed deal: %s", err)
	}
	if len(okDeals) == 0 {
		if len(failedStartedDeals) != 1 {
			return ffs.FilStorage{}, fmt.Errorf("failed started deals must be of size 1, this should never happen")
		}
		fc.l.Log(ctx, "Starting renewal deal proposal failed: %s", failedStartedDeals[0].Message)
		return ffs.FilStorage{}, failedStartedDeals[0]
	}

	var dealError ffs.DealError
	okDeal, err := fc.WaitForDeal(ctx, c, okDeals[0])
	if err != nil && !errors.As(err, &dealError) {
		return ffs.FilStorage{}, ffs.DealError{ProposalCid: c, Message: fmt.Sprintf("waiting for renew deal: %s", err)}
	}
	return okDeal, err
}

// makeDeals starts deals with the specified miners. It returns a slice with all the ProposalCids
// that were started successfully, and a slice of DealError with deals that failed to be started.
func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, size uint64, cfgs []deals.StorageDealConfig, fcfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, error) {
	for _, cfg := range cfgs {
		fc.l.Log(ctx, "Proposing deal to miner %s with %d fil per epoch...", cfg.Miner, cfg.EpochPrice)
	}

	sres, err := fc.dm.Store(ctx, fcfg.Addr, c, size, cfgs, uint64(fcfg.DealMinDuration))
	if err != nil {
		return nil, nil, fmt.Errorf("storing deals in deal module: %s", err)
	}
	var okDeals []cid.Cid
	var failedDeals []ffs.DealError
	for _, r := range sres {
		if !r.Success {
			fc.l.Log(ctx, "Proposal with miner %s failed: %s", r.Config.Miner, r.Message)
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

// WaitForDeal blocks the provided Deal Proposal reaches a final state.
// If the deal finishes successfully it returns a FilStorage result.
// If the deal finished with error, it returns a ffs.DealError error
// result, so it should be considered in error handling.
func (fc *FilCold) WaitForDeal(ctx context.Context, c cid.Cid, proposal cid.Cid) (ffs.FilStorage, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDi, err := fc.dm.Watch(ctx, []cid.Cid{proposal})
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("watching proposals in deals module: %s", err)
	}

	var activeProposal ffs.FilStorage
	for di := range chDi {
		switch di.StateID {
		case storagemarket.StorageDealActive:
			activeProposal = ffs.FilStorage{
				ProposalCid:     di.ProposalCid,
				PieceCid:        di.PieceCID,
				Duration:        int64(di.Duration),
				Miner:           di.Miner,
				ActivationEpoch: di.ActivationEpoch,
				StartEpoch:      di.StartEpoch,
				EpochPrice:      di.PricePerEpoch,
			}
			fc.l.Log(ctx, "Deal %d with miner %s is active on-chain", di.DealID, di.Miner)
			return activeProposal, nil
		case storagemarket.StorageDealError, storagemarket.StorageDealFailing:
			log.Errorf("deal %d failed with state %s: %s", di.DealID, storagemarket.DealStates[di.StateID], di.Message)
			fc.l.Log(ctx, "DealID %d with miner %s failed and won't be active on-chain: %s", di.DealID, di.Miner, di.Message)

			return ffs.FilStorage{}, ffs.DealError{ProposalCid: di.ProposalCid, Miner: di.Miner, Message: di.Message}
		default:
			fc.l.Log(ctx, "Deal with miner %s changed state to %s", di.Miner, storagemarket.DealStates[di.StateID])
		}
	}
	return ffs.FilStorage{}, fmt.Errorf("aborted due to cancellation")
}

// estimatePieceSize estimates the size of the Piece that will be built by the filecoin client
// when making the deal. This calculation should consider the unique node sizes of the DAG, and
// padding. It's important to not underestimate the size since that would lead to deal rejection
// since the miner won't accept the further calculated PricePerEpoch.
func (fc *FilCold) calculatePieceSize(ctx context.Context, c cid.Cid) (uint64, error) {
	// Get unique nodes.
	seen := cid.NewSet()
	if err := dag.Walk(ctx, fc.getLinks, c, seen.Visit); err != nil {
		return 0, fmt.Errorf("walking dag for size calculation: %s", err)
	}

	// Account for CAR header size.
	carHeader := car.CarHeader{
		Roots:   []cid.Cid{c},
		Version: 1,
	}
	totalSize, err := car.HeaderSize(&carHeader)
	if err != nil {
		return 0, fmt.Errorf("calculating car header size: %s", err)
	}

	// Calculate total unique node sizes.
	buf := make([]byte, 8)
	f := func(c cid.Cid) error {
		s, err := fc.ipfs.Block().Stat(ctx, path.IpfsPath(c))
		if err != nil {
			return fmt.Errorf("getting stats from DAG node: %s", err)
		}
		size := uint64(s.Size())
		carBlockHeaderSize := uint64(binary.PutUvarint(buf, size))
		totalSize += carBlockHeaderSize + size
		return nil
	}
	if err := seen.ForEach(f); err != nil {
		return 0, fmt.Errorf("aggregating unique nodes size: %s", err)
	}

	// Consider padding.
	paddedSize := padreader.PaddedSize(totalSize).Padded()

	return uint64(paddedSize), nil
}

func (fc *FilCold) getLinks(ctx context.Context, c cid.Cid) ([]*format.Link, error) {
	return fc.ipfs.Object().Links(ctx, path.IpfsPath(c))
}

func makeDealConfigs(ms ffs.MinerSelector, cntMiners int, f ffs.MinerSelectorFilter, fastRetrieval bool) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetMiners(cntMiners, f)
	if err != nil {
		return nil, fmt.Errorf("getting miners from minerselector: %s", err)
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice, FastRetrieval: fastRetrieval}
	}
	return res, nil
}
