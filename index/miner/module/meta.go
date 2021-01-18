package module

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/powergate/v2/index/miner"
	"github.com/textileio/powergate/v2/iplocation"
	"github.com/textileio/powergate/v2/lotus"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	metaRefreshInterval = time.Hour * 6
	pingTimeout         = time.Second * 5
	pingRateLim         = 1
)

var (
	dsKeyMetaIndex = dsBase.ChildString("meta")
)

func (mi *Index) startMetaWorker(disabled bool) {
	mi.wg.Add(1)
	go func() {
		defer mi.wg.Done()
		if disabled {
			log.Infof("meta index worker disabled")
			return
		}

		for {
			select {
			case <-mi.ctx.Done():
				log.Info("graceful shutdown of meta updater")
				return
			case <-time.After(metaRefreshInterval):
				log.Info("updating meta index...")
				// ToDo: coud have smarter ways of electing which addrs to refresh, and then
				// doing a merge. Will depend if this too slow, but might not be the case
				mi.lock.Lock()
				addrs := make([]string, 0, len(mi.index.OnChain.Miners))
				for addr, miner := range mi.index.OnChain.Miners {
					if miner.Power > 0 {
						addrs = append(addrs, addr)
					}
				}
				mi.lock.Unlock()
				newIndex := updateMetaIndex(mi.ctx, mi.clientBuilder, mi.h, mi.lr, addrs)
				if err := mi.persistMetaIndex(newIndex); err != nil {
					log.Errorf("persisting meta index: %s", err)
				}
				mi.lock.Lock()
				mi.index.Meta = newIndex
				mi.lock.Unlock()
				mi.signaler.Signal() // ToDo: consider a finer-grained signaling
				log.Info("meta index updated")
			}
		}
	}()
}

// updateMetaIndex generates a new index that contains fresh metadata information
// of addrs miners.
func updateMetaIndex(ctx context.Context, clientBuilder lotus.ClientBuilder, h P2PHost, lr iplocation.LocationResolver, addrs []string) miner.MetaIndex {
	client, cls, err := clientBuilder(ctx)
	if err != nil {
		log.Errorf("creating lotus client: %s", err)
		return miner.MetaIndex{}
	}
	defer cls()
	index := miner.MetaIndex{
		Info: make(map[string]miner.Meta),
	}
	rl := make(chan struct{}, pingRateLim)
	var lock sync.Mutex
	for i, a := range addrs {
		if ctx.Err() != nil {
			log.Infof("update meta index canceled")
			return miner.MetaIndex{}
		}
		rl <- struct{}{}
		go func(a string) {
			defer func() { <-rl }()
			si, err := getMeta(ctx, client, h, lr, a)
			if err != nil {
				log.Debugf("error getting static info: %s", err)
				return
			}
			lock.Lock()
			index.Info[a] = merge(index.Info[a], si)
			lock.Unlock()
		}(a)
		if i%100 == 0 {
			stats.Record(context.Background(), mMetaRefreshProgress.M(float64(i)/float64(len(addrs))))
		}
	}
	for i := 0; i < pingRateLim; i++ {
		rl <- struct{}{}
	}
	for _, v := range index.Info {
		if v.Online {
			index.Online++
		}
	}
	index.Offline = uint32(len(addrs)) - index.Online

	stats.Record(context.Background(), mMetaRefreshProgress.M(1))
	ctx, _ = tag.New(context.Background(), tag.Insert(metricOnline, "online"))
	stats.Record(ctx, mMetaPingCount.M(int64(index.Online)))
	ctx, _ = tag.New(context.Background(), tag.Insert(metricOnline, "offline"))
	stats.Record(ctx, mMetaPingCount.M(int64(index.Offline)))

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
func getMeta(ctx context.Context, c *apistruct.FullNodeStruct, h P2PHost, lr iplocation.LocationResolver, straddr string) (miner.Meta, error) {
	si := miner.Meta{
		LastUpdated: time.Now(),
	}
	addr, err := address.NewFromString(straddr)
	if err != nil {
		return si, err
	}
	mi, err := c.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil || mi.PeerId == nil {
		return si, err
	}
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if alive := h.Ping(ctx, *mi.PeerId); !alive {
		return si, fmt.Errorf("peer didn't pong")
	}
	si.Online = true

	if av := h.GetAgentVersion(*mi.PeerId); av != "" {
		si.UserAgent = av
	}

	addrs := h.Addrs(*mi.PeerId)
	if len(addrs) == 0 {
		return si, nil
	}
	if l, err := lr.Resolve(addrs); err == nil {
		si.Location = miner.Location{
			Country:   l.Country,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		}
	}
	return si, nil
}

// persisteMetaIndex saves to datastore a new MetaIndex.
func (mi *Index) persistMetaIndex(index miner.MetaIndex) error {
	buf, err := json.Marshal(index)
	if err != nil {
		return err
	}
	if err := mi.ds.Put(dsKeyMetaIndex, buf); err != nil {
		return err
	}
	return nil
}
