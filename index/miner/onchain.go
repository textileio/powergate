package miner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/textileio/filecoin/lotus/types"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const (
	fullRefreshThreshold = 100
	heightOffset         = uint64(20)
)

// updateOnChainIndex updates on-chain index information when a new tipset
func (mi *MinerIndex) updateOnChainIndex(new *types.TipSet) error {
	log.Info("updating on-chain index...")
	var chainIndex ChainIndex
	newtsk := types.NewTipSetKey(new.Cids...)
	ts, err := mi.store.LoadAndPrune(mi.ctx, newtsk, &chainIndex)
	if err != nil {
		return err
	}
	if chainIndex.Power == nil {
		chainIndex.Power = make(map[string]Power)
	}
	mctx := context.Background()
	start := time.Now()
	log.Infof("current state height %d, new tipset height %d", chainIndex.LastUpdated, new.Height)
	if new.Height-chainIndex.LastUpdated > fullRefreshThreshold {
		mctx, _ = tag.New(mctx, tag.Insert(metricRefreshType, "full"))
		new, err = mi.api.ChainGetTipSetByHeight(mi.ctx, new.Height-heightOffset, new)
		if err != nil {
			return err
		}
		if err := fullRefresh(mi.ctx, mi.api, &chainIndex); err != nil {
			return fmt.Errorf("error doing full refresh: %s", err)
		}

	} else {
		mctx, _ = tag.New(mctx, tag.Insert(metricRefreshType, "delta"))
		if err := deltaRefresh(mi.ctx, mi.api, &chainIndex, *ts, new); err != nil {
			return fmt.Errorf("error doing delta refresh: %s", err)
		}
	}
	chainIndex.LastUpdated = new.Height
	stats.Record(mctx, mRefreshDuration.M(int64(time.Since(start).Milliseconds())))

	mi.lock.Lock()
	mi.index.Chain = chainIndex
	mi.lock.Unlock()

	if err := mi.store.Save(mi.ctx, types.NewTipSetKey(new.Cids...), chainIndex); err != nil {
		return err
	}
	mi.signaler.Signal()
	stats.Record(context.Background(), mUpdatedHeight.M(int64(chainIndex.LastUpdated)))

	return nil
}

// deltaRefresh updates chainIndex information between two TipSet that are on
// the same chain.
func deltaRefresh(ctx context.Context, api API, chainIndex *ChainIndex, fromKey types.TipSetKey, to *types.TipSet) error {
	from, err := api.ChainGetTipSet(ctx, fromKey)
	if err != nil {
		return err
	}
	chg, err := api.StateChangedActors(ctx, from.Blocks[0].ParentStateRoot, to.Blocks[0].ParentStateRoot)
	if err != nil {
		return err
	}
	addrs := make([]string, 0, len(chg))
	for addr := range chg {
		addrs = append(addrs, addr)
	}
	updateForAddrs(ctx, api, chainIndex, addrs)
	return nil
}

// fullRefresh updates chainIndex for all miners information at the currenty
// heaviest tipset.
func fullRefresh(ctx context.Context, api API, chainIndex *ChainIndex) error {
	ts, err := api.ChainHead(ctx)
	if err != nil {
		return err
	}
	addrs, err := api.StateListMiners(ctx, ts)
	if err != nil {
		return err
	}
	updateForAddrs(ctx, api, chainIndex, addrs)
	return nil
}

// updateForAddrs updates chainIndex information for a particular set of addrs.
func updateForAddrs(ctx context.Context, api API, chainIndex *ChainIndex, addrs []string) error {
	var l sync.Mutex
	rl := make(chan struct{}, maxParallelism)
	for i, a := range addrs {
		rl <- struct{}{}
		go func(addr string) {
			defer func() { <-rl }()
			pw, err := getPower(ctx, api, addr)
			if err != nil {
				log.Debug("error getting power: %s", err)
				return
			}
			l.Lock()
			chainIndex.Power[addr] = pw
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
		return fmt.Errorf("cancelled by context")
	default:
	}
	return nil
}

// getPower returns current on-chain power information for a miner
func getPower(ctx context.Context, c API, addr string) (Power, error) {
	mp, err := c.StateMinerPower(ctx, addr, nil)
	if err != nil {
		return Power{}, err
	}
	p := mp.MinerPower.Uint64()
	tp := mp.TotalPower
	return Power{
		Power:    p,
		Relative: float64(p) / float64(tp.Uint64()),
	}, nil
}
