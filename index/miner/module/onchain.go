package module

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/powergate/index/miner"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const (
	// fullThreshold is the maximum height difference between last saved
	// state and new selected tipset that would be a delta-refresh; if greater
	// will do a full refresh.
	fullThreshold = 100
	// hOffset is the # of tipsets from the heaviest chain to
	// consider for index updating; this to reduce sensibility to
	// chain reorgs.
	hOffset = 5
)

// updateOnChainIndex updates on-chain index information in the direction of heaviest tipset
// with some height offset to reduce sensibility to reorgs.
func (mi *Index) updateOnChainIndex() error {
	log.Info("updating on-chain index...")
	heaviest, err := mi.api.ChainHead(mi.ctx)
	if err != nil {
		return err
	}
	if heaviest.Height()-hOffset <= 0 {
		return nil
	}
	new, err := mi.api.ChainGetTipSetByHeight(mi.ctx, heaviest.Height()-hOffset, heaviest.Key())
	if err != nil {
		return err
	}

	var chainIndex miner.ChainIndex
	newtsk := new.Key()
	ts, err := mi.store.LoadAndPrune(mi.ctx, newtsk, &chainIndex)
	if err != nil {
		return fmt.Errorf("error getting last saved from store: %s", err)
	}
	if chainIndex.Miners == nil {
		chainIndex.Miners = make(map[string]miner.OnChainData)
	}
	hdiff := int64(new.Height()) - chainIndex.LastUpdated
	if hdiff == 0 {
		return nil
	}

	mctx := context.Background()
	start := time.Now()
	log.Infof("current state height %d, new tipset height %d", chainIndex.LastUpdated, new.Height())
	if hdiff > fullThreshold || chainIndex.LastUpdated == 0 {
		mctx, _ = tag.New(mctx, tag.Insert(metricRefreshType, "full"))
		if err := fullRefresh(mi.ctx, mi.api, &chainIndex); err != nil {
			return fmt.Errorf("error doing full refresh: %s", err)
		}
	} else {
		mctx, _ = tag.New(mctx, tag.Insert(metricRefreshType, "delta"))
		if err := deltaRefresh(mi.ctx, mi.api, &chainIndex, *ts, new); err != nil {
			return fmt.Errorf("error doing delta refresh: %s", err)
		}
	}
	chainIndex.LastUpdated = int64(new.Height())
	stats.Record(mctx, mRefreshDuration.M(time.Since(start).Milliseconds()))

	mi.lock.Lock()
	mi.index.OnChain = chainIndex
	mi.lock.Unlock()

	if err := mi.store.Save(mi.ctx, new.Key(), chainIndex); err != nil {
		return fmt.Errorf("error when saving state to store: %s", err)
	}
	mi.signaler.Signal()
	stats.Record(context.Background(), mUpdatedHeight.M(chainIndex.LastUpdated))

	return nil
}

// deltaRefresh updates chainIndex information between two TipSet that are on
// the same chain.
func deltaRefresh(ctx context.Context, api *apistruct.FullNodeStruct, chainIndex *miner.ChainIndex, fromKey types.TipSetKey, to *types.TipSet) error {
	from, err := api.ChainGetTipSet(ctx, fromKey)
	if err != nil {
		return err
	}
	chg, err := api.StateChangedActors(ctx, from.Blocks()[0].ParentStateRoot, to.Blocks()[0].ParentStateRoot)
	if err != nil {
		return err
	}
	addrs := make([]address.Address, 0, len(chg))
	for addr := range chg {
		a, err := address.NewFromString(addr)
		if err != nil {
			return err
		}
		addrs = append(addrs, a)
	}
	if err := updateForAddrs(ctx, api, chainIndex, addrs); err != nil {
		return err
	}
	return nil
}

// fullRefresh updates chainIndex for all miners information at the currently
// heaviest tipset.
func fullRefresh(ctx context.Context, api *apistruct.FullNodeStruct, chainIndex *miner.ChainIndex) error {
	ts, err := api.ChainHead(ctx)
	if err != nil {
		return err
	}
	addrs, err := api.StateListMiners(ctx, ts.Key())
	if err != nil {
		return err
	}
	if err := updateForAddrs(ctx, api, chainIndex, addrs); err != nil {
		return fmt.Errorf("updating for addresses: %s", err)
	}
	return nil
}

// updateForAddrs updates chainIndex information for a particular set of addrs.
func updateForAddrs(ctx context.Context, api *apistruct.FullNodeStruct, chainIndex *miner.ChainIndex, addrs []address.Address) error {
	var l sync.Mutex
	rl := make(chan struct{}, maxParallelism)
	for i, a := range addrs {
		rl <- struct{}{}
		go func(addr address.Address) {
			defer func() { <-rl }()
			ocd, err := getOnChainData(ctx, api, addr)
			if err != nil {
				log.Debug("getting power: %s", err)
				return
			}
			l.Lock()
			chainIndex.Miners[addr.String()] = ocd
			l.Unlock()
		}(a)
		stats.Record(context.Background(), mOnChainRefreshProgress.M(float64(i)/float64(len(addrs))))
	}
	for i := 0; i < maxParallelism; i++ {
		rl <- struct{}{}
	}

	stats.Record(context.Background(), mOnChainRefreshProgress.M(1))

	select {
	case <-ctx.Done():
		return fmt.Errorf("canceled by context")
	default:
	}
	return nil
}

func getOnChainData(ctx context.Context, c *apistruct.FullNodeStruct, addr address.Address) (miner.OnChainData, error) {
	// Power of miner.
	mp, err := c.StateMinerPower(ctx, addr, types.EmptyTSK)
	if err != nil {
		return miner.OnChainData{}, fmt.Errorf("getting miner power: %s", err)
	}
	p := mp.MinerPower.RawBytePower.Uint64()

	// Sector size
	info, err := c.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return miner.OnChainData{}, fmt.Errorf("getting sector size: %s", err)
	}

	// Active deals
	sectors, err := c.StateMinerSectors(ctx, addr, nil, false, types.EmptyTSK)
	if err != nil {
		return miner.OnChainData{}, fmt.Errorf("getting sectors: %s", err)
	}
	var activeDeals uint64
	for _, s := range sectors {
		activeDeals += uint64(len(s.Info.DealIDs))
	}
	return miner.OnChainData{
		Power:         p,
		RelativePower: float64(p) / float64(mp.TotalPower.RawBytePower.Uint64()),
		SectorSize:    uint64(info.SectorSize),
		ActiveDeals:   activeDeals,
	}, nil
}
