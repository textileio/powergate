package module

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/util"
)

const (
	defaultDealStartOffset = 48 * 60 * 60 / util.EpochDurationSeconds // 48hs
)

var (
	// ErrDealNotFound indicates a particular ProposalCid from a deal isn't found on-chain. Currently,
	// in Lotus this indicates that it may never existed on-chain, or it existed but it already expired
	// (currEpoch > StartEpoch+Duration).
	ErrDealNotFound = errors.New("deal not found on-chain")
)

// Store create Deal Proposals with all miners indicated in dcfgs. The epoch price
// is automatically calculated considering each miner epoch price and piece size.
// The data of dataCid should be already imported to the Filecoin Client or should be
// accessible to it. (e.g: is integrated with an IPFS node).
func (m *Module) Store(ctx context.Context, waddr string, dataCid cid.Cid, dataSize int64, pieceSize abi.PaddedPieceSize, pieceCid cid.Cid, dcfgs []deals.StorageDealConfig, minDuration uint64) ([]deals.StoreResult, error) {
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
			VerifiedDeal:      c.VerifiedDeal,
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
		m.recordDeal(params, *p, dataSize)
	}

	return res, nil
}

// Watch returns a channel with state changes of indicated proposals.
func (m *Module) Watch(ctx context.Context, proposal cid.Cid) (<-chan deals.StorageDealInfo, error) {
	updates := make(chan deals.StorageDealInfo)

	go func() {
		defer close(updates)

		watcherUpdates := make(chan struct{}, 20)
		if err := m.dealWatcher.Subscribe(watcherUpdates, proposal); err != nil {
			log.Errorf("subscribing to deal-watcher channel: %s", err)
		}
		defer func() {
			if err := m.dealWatcher.Unsubscribe(watcherUpdates, proposal); err != nil {
				log.Errorf("unregistering from deal-watcher: %s", err)
			}
		}()

		// Notify once so that subscribers get a result quickly
		last, err := m.getStorageDealInfo(ctx, proposal)
		if err != nil {
			log.Errorf("notifying latest proposal status: %s", err)
			return
		}
		updates <- last

		// Then notify every m.pollDuration
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.pollDuration):
			case <-watcherUpdates:
			}

			sdi, err := m.getStorageDealInfo(ctx, proposal)
			if err != nil {
				log.Errorf("notifying latests proposal status: %s", err)
				return
			}
			if last.StateID != sdi.StateID {
				last = sdi
				updates <- last
			}
		}
	}()

	return updates, nil
}

func (m *Module) getStorageDealInfo(ctx context.Context, proposal cid.Cid) (deals.StorageDealInfo, error) {
	lapi, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return deals.StorageDealInfo{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	dinfo, err := robustClientGetDealInfo(ctx, lapi, proposal)
	if err != nil {
		return deals.StorageDealInfo{}, fmt.Errorf("getting deal proposal info %s: %s", proposal, err)
	}

	sdi, err := fromLotusDealInfo(ctx, lapi, dinfo)
	if err != nil {
		return deals.StorageDealInfo{}, fmt.Errorf("converting proposal cid %s from lotus deal info: %v", util.CidToString(proposal), err)
	}

	return sdi, nil
}

func fromLotusDealInfo(ctx context.Context, client *api.FullNodeStruct, dinfo *api.DealInfo) (deals.StorageDealInfo, error) {
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
