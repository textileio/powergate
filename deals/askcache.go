package deals

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/filecoin/lotus/types"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	queryAskRateLim  = 25
	queryAskTimeout  = time.Second * 20
	dsStorageAskBase = datastore.NewKey("/deals/storageask")
)

// Query specifies filtering and paging data to retrieve active Asks
type Query struct {
	MaxPrice  uint64
	PieceSize uint64
	Limit     int
	Offset    int
}

// StorageAsk has information about an active ask from a storage miner
type StorageAsk struct {
	Price        uint64
	MinPieceSize uint64
	Miner        string
	Timestamp    uint64
	Expiry       uint64
}

// AvailableAsk executes a query to retrieve active Asks
func (d *DealModule) AvailableAsks(q Query) ([]StorageAsk, error) {
	d.askCacheLock.RLock()
	defer d.askCacheLock.RUnlock()
	var res []StorageAsk
	offset := q.Offset
	for _, sa := range d.askCache {
		if q.MaxPrice != 0 && types.BigCmp(sa.Price, types.NewInt(q.MaxPrice)) == 1 {
			break
		}
		if q.PieceSize != 0 && sa.MinPieceSize > q.PieceSize {
			continue
		}
		if offset > 0 {
			offset--
			continue
		}
		res = append(res, StorageAsk{
			Price:        sa.Price.Uint64(),
			MinPieceSize: sa.MinPieceSize,
			Miner:        sa.Miner,
			Timestamp:    sa.Timestamp,
			Expiry:       sa.Expiry,
		})
		if q.Limit != 0 && len(res) == q.Limit {
			break
		}
	}
	return res, nil
}

func (d *DealModule) runBackgroundAskCache() {
	defer close(d.closed)
	if err := d.updateMinerAsks(); err != nil {
		log.Errorf("error when updating miners asks: %s", err)
	}
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(askRefreshInterval):
			log.Debug("refreshing ask cache")
			if err := d.updateMinerAsks(); err != nil {
				log.Errorf("error when updating miners asks: %s", err)
			}
		}
	}
}

func (d *DealModule) updateMinerAsks() error {
	asks, err := takeFreshAskSnapshot(d.ctx, d.api)
	if err != nil {
		return err
	}

	sort.Slice(asks, func(i, j int) bool {
		return types.BigCmp(asks[i].Price, asks[j].Price) == -1
	})

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for _, ask := range asks {
		buf.Reset()
		if err := encoder.Encode(ask); err != nil {
			log.Errorf("error when marshaling storage ask: %s", err)
			return err
		}
		if err := d.ds.Put(dsStorageAskBase.ChildString(ask.Miner), buf.Bytes()); err != nil {
			log.Errorf("error when persiting storage ask: %s", err)
			return err
		}
	}
	d.askCacheLock.Lock()
	d.askCache = asks
	d.askCacheLock.Unlock()

	return nil
}

func takeFreshAskSnapshot(baseCtx context.Context, api DealerAPI) ([]*types.StorageAsk, error) {
	startTime := time.Now()
	defer stats.Record(context.Background(), mAskCacheFullRefreshTime.M(startTime.Sub(time.Now()).Milliseconds()))

	ctx, cancel := context.WithTimeout(baseCtx, time.Second*5)
	defer cancel()
	rateLim := make(chan struct{}, queryAskRateLim)
	addrs, err := api.StateListMiners(ctx, nil)
	if err != nil {
		return nil, err
	}
	stats.Record(context.Background(), mMinerCount.M(int64(len(addrs))))

	var wg sync.WaitGroup
	askCh := make(chan *types.StorageAsk)
	for _, a := range addrs {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			rateLim <- struct{}{}
			defer func() { <-rateLim }()
			ctx, cancel := context.WithTimeout(baseCtx, queryAskTimeout)
			defer cancel()
			pid, err := api.StateMinerPeerID(ctx, a, nil)
			if err != nil {
				log.Info("error getting pid of %s: %s", a, err)
				return
			}

			ask, err := api.ClientQueryAsk(ctx, pid, a)
			if err != nil {
				return
			}

			askCh <- ask.Ask
		}(a)
	}
	go func() {
		wg.Wait()
		close(askCh)
	}()
	asks := make([]*types.StorageAsk, 0, len(addrs))
	for sa := range askCh {
		asks = append(asks, sa)
	}

	select {
	case <-baseCtx.Done():
		return nil, fmt.Errorf("cancelled on request")
	default:
	}

	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "FAIL"))
	stats.Record(ctx, mAskQuery.M(int64(len(addrs)-len(asks))))
	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "OK"))
	stats.Record(ctx, mAskQuery.M(int64(len(asks))))

	return asks, nil
}
