package lotusidx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/v2/index/miner"
	"github.com/textileio/powergate/v2/iplocation"
	"github.com/textileio/powergate/v2/lotus"
)

var (
	metaRefreshInterval = time.Hour * 6
	metaRateLim         = 1
)

func (mi *Index) startMetaWorker() {
	mi.wg.Add(1)

	startRun := make(chan struct{}, 1)

	if mi.conf.RefreshOnStart {
		startRun <- struct{}{}
	}

	go func() {
		defer mi.wg.Done()

		for {
			select {
			case <-mi.ctx.Done():
				log.Info("graceful shutdown of meta updater")
				return
			case <-time.After(metaRefreshInterval):
				if err := mi.tickUpdateMetaIndex(); err != nil {
					log.Error(err)
				}
			case <-startRun:
				if err := mi.tickUpdateMetaIndex(); err != nil {
					log.Error(err)
				}
			}
		}
	}()
}

func (mi *Index) tickUpdateMetaIndex() error {
	start := time.Now()
	log.Info("updating meta index...")
	defer log.Info("meta index updated")

	mi.lock.Lock()
	addrs := make([]string, 0, len(mi.index.OnChain.Miners))
	for addr, miner := range mi.index.OnChain.Miners {
		if miner.Power > 0 {
			addrs = append(addrs, addr)
		}
	}
	mi.lock.Unlock()

	newIndex := mi.updateMetaIndex(mi.ctx, mi.cb, addrs)
	if err := mi.store.SaveMetadata(newIndex); err != nil {
		return fmt.Errorf("persisting meta index: %s", err)
	}

	mi.lock.Lock()
	mi.index.Meta = newIndex
	mi.lock.Unlock()

	mi.signaler.Signal()

	mi.meterRefreshDuration.Record(context.Background(), time.Since(start).Milliseconds(), metaSubindex)

	return nil
}

// updateMetaIndex generates a new index that contains fresh metadata information
// of addrs miners.
func (mi *Index) updateMetaIndex(ctx context.Context, clientBuilder lotus.ClientBuilder, addrs []string) miner.MetaIndex {
	client, cls, err := clientBuilder(ctx)
	if err != nil {
		log.Errorf("creating lotus client: %s", err)
		return miner.MetaIndex{}
	}
	defer cls()
	index := miner.MetaIndex{
		Info: make(map[string]miner.Meta),
	}
	rl := make(chan struct{}, metaRateLim)
	var lock sync.Mutex
	for i, a := range addrs {
		if ctx.Err() != nil {
			log.Infof("update meta index canceled")
			return miner.MetaIndex{}
		}
		rl <- struct{}{}
		go func(a string) {
			defer func() { <-rl }()
			si, err := getMeta(ctx, client, mi.h, mi.lr, a)
			if err != nil {
				log.Debugf("getting static info: %s", err)
				return
			}
			lock.Lock()
			index.Info[a] = merge(index.Info[a], si)
			lock.Unlock()
		}(a)
		mi.metricLock.Lock()
		mi.metaProgress = float64(i) / float64(len(addrs))
		mi.metricLock.Unlock()
	}
	for i := 0; i < metaRateLim; i++ {
		rl <- struct{}{}
	}
	mi.metricLock.Lock()
	mi.metaProgress = float64(1)
	mi.metricLock.Unlock()

	return index
}

func merge(old miner.Meta, upt miner.Meta) miner.Meta {
	if upt.Location.Country == "" {
		upt.Location.Country = old.Location.Country
	}

	if upt.Location.Latitude == 0 {
		upt.Location.Latitude = old.Location.Latitude
	}

	if upt.Location.Longitude == 0 {
		upt.Location.Longitude = old.Location.Longitude
	}

	return upt
}

// getMeta returns fresh metadata information about a miner.
func getMeta(ctx context.Context, c *api.FullNodeStruct, h P2PHost, lr iplocation.LocationResolver, straddr string) (miner.Meta, error) {
	si := miner.Meta{
		LastUpdated: time.Now(),
	}
	addr, err := address.NewFromString(straddr)
	if err != nil {
		return si, err
	}
	mi, err := c.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return si, err
	}

	if mi.PeerId != nil {
		if av := h.GetAgentVersion(*mi.PeerId); av != "" {
			si.UserAgent = av
		}
	}

	if len(mi.Multiaddrs) == 0 {
		return si, nil
	}
	var maddrs []multiaddr.Multiaddr
	for _, ma := range mi.Multiaddrs {
		pma, err := multiaddr.NewMultiaddrBytes(ma)
		if err != nil {
			continue
		}
		maddrs = append(maddrs, pma)
	}
	if l, err := lr.Resolve(maddrs); err == nil {
		si.Location = miner.Location{
			Country:   l.Country,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		}
	}
	return si, nil
}
