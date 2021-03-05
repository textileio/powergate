package lotusidx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/powergate/v2/index/miner"
)

func (mi *Index) updateOnChainIndex(ctx context.Context) error {
	log.Info("updating on-chain index...")
	defer log.Info("on-chain index updated")

	client, cls, err := mi.cb(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	start := time.Now()
	addrs, err := client.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("calling state-list-miners: %s", err)
	}
	log.Infof("listing all on-chain miners took %.2f seconds", time.Since(start).Seconds())

	newIndex := miner.ChainIndex{
		Miners: map[string]miner.OnChainMinerData{},
	}
	if err := updateForAddrs(ctx, client, &newIndex, addrs); err != nil {
		return fmt.Errorf("updating for addresses: %s", err)
	}
	chainHead, err := client.ChainHead(ctx)
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	newIndex.LastUpdated = int64(chainHead.Height())

	mi.lock.Lock()
	mi.index.OnChain = newIndex
	mi.lock.Unlock()

	if err := mi.store.SaveOnChain(ctx, newIndex); err != nil {
		return fmt.Errorf("saving on-chain index to store: %s", err)
	}
	mi.signaler.Signal()

	return nil
}

// updateForAddrs updates chainIndex information for a particular set of addrs.
func updateForAddrs(ctx context.Context, api *apistruct.FullNodeStruct, chainIndex *miner.ChainIndex, addrs []address.Address) error {
	var l sync.Mutex
	rl := make(chan struct{}, maxParallelism)
	for i, a := range addrs {
		if ctx.Err() != nil {
			return fmt.Errorf("update on-chain index canceled")
		}
		rl <- struct{}{}
		go func(addr address.Address) {
			defer func() { <-rl }()
			ocd, err := getOnChainData(ctx, api, addr)
			if err != nil {
				log.Debugf("getting onchain data: %s", err)
				return
			}
			l.Lock()
			chainIndex.Miners[addr.String()] = ocd
			l.Unlock()
		}(a)
		if i%10000 == 0 {
			log.Infof("on-chain idx progress %d/%d", i, len(addrs))
		}
	}
	for i := 0; i < maxParallelism; i++ {
		rl <- struct{}{}
	}

	if ctx.Err() != nil {
		return fmt.Errorf("update on-chain index canceled")
	}

	return nil
}

func getOnChainData(ctx context.Context, c *apistruct.FullNodeStruct, addr address.Address) (miner.OnChainMinerData, error) {
	// Power of miner.
	mp, err := c.StateMinerPower(ctx, addr, types.EmptyTSK)
	if err != nil {
		return miner.OnChainMinerData{}, fmt.Errorf("getting miner power: %s", err)
	}

	// Sector size
	info, err := c.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return miner.OnChainMinerData{}, fmt.Errorf("getting sector size: %s", err)
	}

	// Sectors
	sectors, err := c.StateMinerSectorCount(ctx, addr, types.EmptyTSK)
	if err != nil {
		return miner.OnChainMinerData{}, fmt.Errorf("getting sectors count: %s", err)
	}

	p := mp.MinerPower.RawBytePower.Uint64()
	return miner.OnChainMinerData{
		Power:         p,
		RelativePower: float64(p) / float64(mp.TotalPower.RawBytePower.Uint64()),
		SectorSize:    uint64(info.SectorSize),
		SectorsLive:   sectors.Live,
		SectorsActive: sectors.Active,
		SectorsFaulty: sectors.Faulty,
	}, nil
}
