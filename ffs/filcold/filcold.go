package filcold

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/deals/module"
	dealsModule "github.com/textileio/powergate/v2/deals/module"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/util"
)

const (
	unsyncedThreshold = 10
)

var (
	log = logger.Logger("ffs-filcold")
)

// FilCold is a ColdStorage implementation which saves data in the Filecoin network.
// It assumes the underlying Filecoin client has access to an IPFS node where data is stored.
type FilCold struct {
	ms             ffs.MinerSelector
	dm             *dealsModule.Module
	ipfs           iface.CoreAPI
	chain          FilChain
	l              ffs.JobLogger
	lsm            *lotus.SyncMonitor
	minPieceSize   uint64
	semaphDealPrep chan struct{}
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// FilChain is an abstraction of a Filecoin node to get information of the network.
type FilChain interface {
	GetHeight(context.Context) (uint64, error)
}

// New returns a new FilCold instance.
func New(ms ffs.MinerSelector, dm *dealsModule.Module, ipfs iface.CoreAPI, chain FilChain, l ffs.JobLogger, lsm *lotus.SyncMonitor, minPieceSize uint64, maxParallelDealPreparing int) *FilCold {
	return &FilCold{
		ms:             ms,
		dm:             dm,
		ipfs:           ipfs,
		chain:          chain,
		l:              l,
		lsm:            lsm,
		minPieceSize:   minPieceSize,
		semaphDealPrep: make(chan struct{}, maxParallelDealPreparing),
	}
}

// Fetch fetches the stored Cid data.The data will be considered available
// to the underlying blockstore.
func (fc *FilCold) Fetch(ctx context.Context, pyCid cid.Cid, piCid *cid.Cid, waddr string, miners []string, maxPrice uint64, selector string) (ffs.FetchInfo, error) {
	miner, events, err := fc.dm.Fetch(ctx, waddr, pyCid, piCid, miners)
	if err != nil {
		return ffs.FetchInfo{}, fmt.Errorf("fetching from deal module: %s", err)
	}
	fc.l.Log(ctx, "Fetching from %s...", miner)
	var fundsSpent uint64
	var lastMsg string
	for e := range events {
		if e.Err != "" {
			return ffs.FetchInfo{}, fmt.Errorf("event error in retrieval progress: %s", e.Err)
		}
		strEvent := retrievalmarket.ClientEvents[e.Event]
		strDealStatus := retrievalmarket.DealStatuses[e.Status]
		fundsSpent = e.FundsSpent.Uint64()
		newMsg := fmt.Sprintf("Received %.2fGiB, total spent: %sFIL (%s/%s)", float64(e.BytesReceived)/1024/1024/1024, util.AttoFilToFil(fundsSpent), strEvent, strDealStatus)
		if newMsg != lastMsg {
			fc.l.Log(ctx, newMsg)
			lastMsg = newMsg
		}
	}
	return ffs.FetchInfo{RetrievedMiner: miner, FundsSpent: fundsSpent}, nil
}

func (fc *FilCold) calculateDealPiece(ctx context.Context, c cid.Cid) (abi.PaddedPieceSize, cid.Cid, error) {
	fc.l.Log(ctx, "Entering deal preprocessing queue...")
	select {
	case fc.semaphDealPrep <- struct{}{}:
	case <-ctx.Done():
		return 0, cid.Undef, fmt.Errorf("canceled by context")
	}
	defer func() { <-fc.semaphDealPrep }()
	for {
		if fc.lsm.SyncHeightDiff() < unsyncedThreshold {
			break
		}
		log.Warnf("backpressure from unsynced Lotus node")
		select {
		case <-ctx.Done():
			return 0, cid.Undef, fmt.Errorf("canceled by context")
		case <-time.After(time.Minute):
		}
	}
	fc.l.Log(ctx, "Calculating piece size...")
	piece, err := fc.dm.CalculateDealPiece(ctx, c)
	if err != nil {
		return 0, cid.Undef, fmt.Errorf("getting cid cummulative size: %s", err)
	}
	return piece.PieceSize, piece.PieceCID, nil
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation. It returns a slice of ProposalCids that were correctly
// started, and a slice of with Proposal Cids rejected. Returned proposed deals can be tracked
// with the WaitForDeal API.
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, cfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, abi.PaddedPieceSize, error) {
	pieceSize, pieceCid, err := fc.calculateDealPiece(ctx, c)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("getting cid cummulative size: %s", err)
	}
	fc.l.Log(ctx, "Calculated piece size is %d MiB.", pieceSize/1024/1024)

	if uint64(pieceSize) < fc.minPieceSize {
		return nil, nil, 0, fmt.Errorf("Piece size is below allowed minimum %d MiB", fc.minPieceSize/1024/1024)
	}
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: cfg.ExcludedMiners,
		CountryCodes:   cfg.CountryCodes,
		TrustedMiners:  cfg.TrustedMiners,
		MaxPrice:       cfg.MaxPrice,
		PieceSize:      uint64(pieceSize),
	}
	cfgs, err := makeDealConfigs(fc.ms, cfg.RepFactor, f, cfg.FastRetrieval, cfg.DealStartOffset)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("making deal configs: %s", err)
	}

	okDeals, failedStartingDeals, err := fc.makeDeals(ctx, c, pieceSize, pieceCid, cfgs, cfg)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("starting deals: %s", err)
	}
	return okDeals, failedStartingDeals, pieceSize, nil
}

// GetDealInfo returns on-chain information for a deal.
func (fc *FilCold) GetDealInfo(ctx context.Context, dealID uint64) (api.MarketDeal, error) {
	di, err := fc.dm.GetDealInfo(ctx, dealID)
	if err == module.ErrDealNotFound {
		return api.MarketDeal{}, ffs.ErrOnChainDealNotFound
	}
	if err != nil {
		return api.MarketDeal{}, fmt.Errorf("getting deal information: %s", err)
	}
	if di.State.SlashEpoch != -1 {
		return api.MarketDeal{}, ffs.ErrOnChainDealNotFound
	}

	return di, nil
}

// EnsureRenewals analyzes a FilInfo state for a Cid and executes renewals considering the FilConfig desired configuration.
// Deal status updates are sent on the provided dealUpdates channel.
// The caller should close the channel once all calls to EnsureRenewals have returned.
// It returns an updated FilInfo for the Cid. All prevous Proposals in the received FilInfo are kept, only flagging the ones
// that got renewed with Renewed=true. New deals from renewals are added to the returned FilInfo.
// Note: Most probably all this code should change in the future, when Filecoin supports telling the miner which deal is about to
// expire that we're interested in extending the deal duration. Now we should make a new deal from scratch (send data, etc).
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, cfg ffs.FilConfig, dealFinalityTimeout time.Duration, dealUpdates chan deals.StorageDealInfo) (ffs.FilInfo, []ffs.DealError, error) {
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

	toRenew := renewable[:numToBeRenewed]
	var newDealErrors []ffs.DealError
	for i, p := range toRenew {
		var dealError ffs.DealError
		newProposal, err := fc.renewDeal(ctx, c, abi.PaddedPieceSize(inf.Size), p.PieceCid, p, cfg, dealFinalityTimeout, dealUpdates)
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

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, pieceSize abi.PaddedPieceSize, pieceCid cid.Cid, p ffs.FilStorage, fcfg ffs.FilConfig, waitDealTimeout time.Duration, dealUpdates chan deals.StorageDealInfo) (ffs.FilStorage, error) {
	f := ffs.MinerSelectorFilter{
		ExcludedMiners: fcfg.ExcludedMiners,
		CountryCodes:   fcfg.CountryCodes,
		TrustedMiners:  []string{p.Miner},
		MaxPrice:       fcfg.MaxPrice,
		PieceSize:      uint64(pieceSize),
	}
	dealConfig, err := makeDealConfigs(fc.ms, 1, f, fcfg.FastRetrieval, fcfg.DealStartOffset)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("making new deal config: %s", err)
	}

	okDeals, failedStartedDeals, err := fc.makeDeals(ctx, c, pieceSize, pieceCid, dealConfig, fcfg)
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
	okDeal, err := fc.WaitForDeal(ctx, c, okDeals[0], waitDealTimeout, dealUpdates)
	if err != nil && !errors.As(err, &dealError) {
		return ffs.FilStorage{}, ffs.DealError{ProposalCid: c, Message: fmt.Sprintf("waiting for renew deal: %s", err)}
	}
	return okDeal, err
}

// makeDeals starts deals with the specified miners. It returns a slice with all the ProposalCids
// that were started successfully, and a slice of DealError with deals that failed to be started.
func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, pieceSize abi.PaddedPieceSize, pieceCid cid.Cid, cfgs []deals.StorageDealConfig, fcfg ffs.FilConfig) ([]cid.Cid, []ffs.DealError, error) {
	for {
		if fc.lsm.SyncHeightDiff() < unsyncedThreshold {
			break
		}
		log.Warnf("lotus backpressure from unsynced node")
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("canceled by context")
		case <-time.After(time.Minute):
		}
	}

	for _, cfg := range cfgs {
		fc.l.Log(ctx, "Proposing deal to miner %s with %s FIL per epoch...", cfg.Miner, util.AttoFilToFil(cfg.EpochPrice))
	}

	sres, err := fc.dm.Store(ctx, fcfg.Addr, c, pieceSize, pieceCid, cfgs, uint64(fcfg.DealMinDuration))
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
// Deal status updates are sent on the provided dealUpdates channel.
// The caller should close the channel once all calls to WaitForDeal have returned.
// If the deal finishes successfully it returns a FilStorage result.
// If the deal finished with error, it returns a ffs.DealError error
// result, so it should be considered in error handling.
func (fc *FilCold) WaitForDeal(ctx context.Context, c cid.Cid, proposal cid.Cid, timeout time.Duration, dealUpdates chan deals.StorageDealInfo) (ffs.FilStorage, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDi, err := fc.dm.Watch(ctx, proposal)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("watching proposals in deals module: %s", err)
	}

	var last deals.StorageDealInfo
Loop:
	for {
		select {
		case <-time.After(timeout):
			msg := fmt.Sprintf("DealID %d with miner %s tracking timed out after waiting for %.0f hours.", last.DealID, last.Miner, timeout.Hours())
			fc.l.Log(ctx, msg)
			return ffs.FilStorage{}, ffs.DealError{ProposalCid: proposal, Miner: last.Miner, Message: msg}
		case di, ok := <-chDi:
			if !ok {
				break Loop
			}
			last = di
			select {
			case dealUpdates <- di:
			default:
				log.Warnf("slow receiver for deal updates for %s", c)
			}
			switch di.StateID {
			case storagemarket.StorageDealActive:
				activeProposal := ffs.FilStorage{
					DealID:     di.DealID,
					PieceCid:   di.PieceCID,
					Duration:   int64(di.Duration),
					Miner:      di.Miner,
					StartEpoch: di.StartEpoch,
					EpochPrice: di.PricePerEpoch,
				}
				fc.l.Log(ctx, "Deal %d with miner %s is active on-chain", di.DealID, di.Miner)

				return activeProposal, nil
			case storagemarket.StorageDealError, storagemarket.StorageDealFailing:
				log.Errorf("deal %d & proposal %s failed with state %s: %s", di.DealID, proposal, storagemarket.DealStates[di.StateID], di.Message)
				fc.l.Log(ctx, "DealID %d with miner %s failed and won't be active on-chain: %s", di.DealID, di.Miner, di.Message)

				return ffs.FilStorage{}, ffs.DealError{ProposalCid: di.ProposalCid, Miner: di.Miner, Message: di.Message}
			default:
				if di.DealID != 0 {
					fc.l.Log(ctx, "Deal %d with miner %s changed state to %s", di.DealID, di.Miner, storagemarket.DealStates[di.StateID])
				} else {
					fc.l.Log(ctx, "Deal with miner %s changed state to %s", di.Miner, storagemarket.DealStates[di.StateID])
				}
			}
		}
	}
	return ffs.FilStorage{}, fmt.Errorf("aborted due to cancellation")
}

func makeDealConfigs(ms ffs.MinerSelector, cntMiners int, f ffs.MinerSelectorFilter, fastRetrieval bool, dealStartOffset int64) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetMiners(cntMiners, f)
	if err != nil {
		return nil, fmt.Errorf("getting miners from minerselector: %s", err)
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{
			Miner:           m.Addr,
			EpochPrice:      m.EpochPrice,
			FastRetrieval:   fastRetrieval,
			DealStartOffset: dealStartOffset,
		}
	}
	return res, nil
}
