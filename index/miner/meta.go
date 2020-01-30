package miner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/textileio/filecoin/fchost"
	"github.com/textileio/filecoin/iplocation"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const (
	metadataRefreshInterval = time.Second * 45
	pingTimeout             = time.Second * 3
	pingRateLim             = 100
)

var (
	dsKeyMetaIndex = dsBase.ChildString("meta")
)

// metaWorker makes a pass on refreshing metadata information about known miners
func (mi *MinerIndex) metaWorker() {
	defer func() { mi.finished <- struct{}{} }()
	mi.chMeta <- struct{}{}
	for {
		select {
		case <-mi.ctx.Done():
			log.Info("graceful shutdown of meta updater")
			return
		case _, ok := <-mi.chMeta:
			if !ok {
				log.Info("meta worker channel closed")
				return
			}
			log.Info("updating meta index...")
			// ToDo: coud have smarter ways of electing which addrs to refresh, and then
			// doing a merge. Will depend if this too slow, but might not be the case
			mi.lock.Lock()
			addrs := make([]string, 0, len(mi.index.Chain.Power))
			for addr := range mi.index.Chain.Power {
				addrs = append(addrs, addr)
			}
			mi.lock.Unlock()
			newIndex, err := updateMetaIndex(mi.ctx, mi.api, mi.h, mi.lr, addrs)
			if err != nil {
				log.Errorf("error when updating meta index: %s", err)
				break
			}
			if err := mi.persistMetaIndex(newIndex); err != nil {
				log.Errorf("error when persisting meta index: %s", err)
			}
			mi.lock.Lock()
			mi.index.Meta = newIndex
			mi.lock.Unlock()
			mi.signaler.Signal() // ToDo: consider a finer-grained signaling
		}
	}
}

// updateMetaIndex generates a new index that contains fresh metadata information
// of addrs miners.
func updateMetaIndex(ctx context.Context, api API, h *fchost.FilecoinHost, lr iplocation.LocationResolver, addrs []string) (MetaIndex, error) {
	index := MetaIndex{
		Info: make(map[string]Meta),
	}
	rl := make(chan struct{}, pingRateLim)
	var lock sync.Mutex
	for i, a := range addrs {
		rl <- struct{}{}
		go func(a string) {
			defer func() { <-rl }()
			si, err := getMeta(ctx, api, h, lr, a)
			if err != nil {
				log.Debugf("error getting static info: %s", err)
				return
			}
			lock.Lock()
			index.Info[a] = si
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

	return index, nil
}

// getMeta returns fresh metadata information about a miner
func getMeta(ctx context.Context, c API, h *fchost.FilecoinHost, lr iplocation.LocationResolver, straddr string) (Meta, error) {
	si := Meta{
		LastUpdated: time.Now(),
	}
	addr, err := address.NewFromString(straddr)
	if err != nil {
		return si, err
	}
	pid, err := c.StateMinerPeerID(ctx, addr, nil)
	if err != nil {
		return si, err
	}
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if alive := h.Ping(ctx, pid); !alive {
		return si, fmt.Errorf("peer didn't pong")
	}
	si.Online = true

	if av := h.GetAgentVersion(pid); av != "" {
		si.UserAgent = av
	}

	addrs := h.Addrs(pid)
	if len(addrs) == 0 {
		return si, nil
	}
	if l, err := lr.Resolve(addrs); err == nil {
		si.Location = Location{
			Country:   l.Country,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		}
	}
	return si, nil
}

// persisteMetaIndex saves to datastore a new MetaIndex
func (mi *MinerIndex) persistMetaIndex(index MetaIndex) error {
	buf, err := cbor.DumpObject(index)
	if err != nil {
		return err
	}
	if err := mi.ds.Put(dsKeyMetaIndex, buf); err != nil {
		return err
	}
	return nil
}
