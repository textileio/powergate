package ask

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/prometheus/common/log"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	queryAskRateLim    = 25
	queryAskTimeout    = time.Second * 20
	askRefreshInterval = time.Second * 10
	dsStorageAskBase   = datastore.NewKey("/index/ask")
)

type AskIndex struct {
	signaler.Signaler
	c  API
	ds datastore.TxnDatastore

	cache []*types.StorageAsk

	lock     sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

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

// API interacts with a Filecoin full-node
type API interface {
	StateListMiners(context.Context, *types.TipSet) ([]string, error)
	ClientQueryAsk(ctx context.Context, p peer.ID, miner string) (*types.SignedStorageAsk, error)
	StateMinerPeerID(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error)
}

func New(ds datastore.TxnDatastore, c API) *AskIndex {
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	ai := &AskIndex{
		c:  c,
		ds: ds,

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go ai.runBackgroundAskCache()
	return ai
}

// AvailableAsks executes a query to retrieve active Asks
func (d *AskIndex) AvailableAsks(q Query) ([]StorageAsk, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var res []StorageAsk
	offset := q.Offset
	for _, sa := range d.cache {
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

func (d *AskIndex) runBackgroundAskCache() {
	defer close(d.finished)
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

func (d *AskIndex) updateMinerAsks() error {
	asks, err := takeFreshAskSnapshot(d.ctx, d.c)
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
	d.lock.Lock()
	d.cache = asks
	d.lock.Unlock()

	return nil
}

func takeFreshAskSnapshot(baseCtx context.Context, api API) ([]*types.StorageAsk, error) {
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
				log.Infof("error getting pid of %s: %s", a, err)
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

func (d *AskIndex) Close() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.closed {
		return
	}
	d.cancel()
	<-d.finished
	d.closed = true
}
