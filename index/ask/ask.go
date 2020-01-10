package ask

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	queryAskRateLim    = 200
	queryAskTimeout    = time.Second * 20
	askRefreshInterval = time.Minute * 30
	dsStorageAskBase   = datastore.NewKey("/index/ask")

	log = logging.Logger("index-ask")
)

func init() {
	cbor.RegisterCborType(StorageAsk{})
}

type AskIndex struct {
	signaler.Signaler
	c  API
	ds datastore.TxnDatastore

	lock       sync.Mutex
	index      Index
	queryCache []*StorageAsk

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

type Index struct {
	LastUpdated time.Time
	MedianPrice uint64
	Miners      map[string]StorageAsk
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
	Miner        string
	Price        uint64
	MinPieceSize uint64
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
		c:        c,
		ds:       ds,
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go ai.runBackgroundAskCache()
	return ai
}

func (d *AskIndex) Get() Index {
	d.lock.Lock()
	defer d.lock.Unlock()
	ii := Index{
		LastUpdated: d.index.LastUpdated,
		Miners:      make(map[string]StorageAsk, len(d.index.Miners)),
	}
	for addr, v := range d.index.Miners {
		ii.Miners[addr] = v
	}
	return ii
}

// Query executes a query to retrieve active Asks
func (ai *AskIndex) Query(q Query) ([]StorageAsk, error) {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	var res []StorageAsk
	offset := q.Offset
	for _, sa := range ai.queryCache {
		if q.MaxPrice != 0 && sa.Price > q.MaxPrice {
			break
		}
		if q.PieceSize != 0 && sa.MinPieceSize > q.PieceSize {
			continue
		}
		if offset > 0 {
			offset--
			continue
		}
		res = append(res, *sa)
		if q.Limit != 0 && len(res) == q.Limit {
			break
		}
	}
	return res, nil
}

func (ai *AskIndex) runBackgroundAskCache() {
	defer close(ai.finished)
	if err := ai.refreshAsks(); err != nil {
		log.Errorf("error when updating miners asks: %s", err)
	}
	for {
		select {
		case <-ai.ctx.Done():
			return
		case <-time.After(askRefreshInterval):
			log.Debug("refreshing ask cache")
			if err := ai.refreshAsks(); err != nil {
				log.Errorf("error when updating miners asks: %s", err)
			}
		}
	}
}

func (ai *AskIndex) refreshAsks() error {
	startTime := time.Now()
	if err := ai.updateIndex(); err != nil {
		return err
	}

	ai.lock.Lock()
	cache := make([]*StorageAsk, 0, len(ai.index.Miners))
	for _, v := range ai.index.Miners {
		cache = append(cache, &v)
	}
	sort.Slice(cache, func(i, j int) bool {
		return cache[i].Price < cache[j].Price
	})
	for _, ask := range ai.index.Miners {
		b, err := cbor.DumpObject(ask)
		if err != nil {
			log.Errorf("error when marshaling storage ask: %s", err)
			return err
		}
		if err := ai.ds.Put(dsStorageAskBase.ChildString(ask.Miner), b); err != nil {
			log.Errorf("error when persiting storage ask: %s", err)
			return err
		}
	}
	ai.lock.Unlock()

	stats.Record(context.Background(), mAskCacheFullRefreshTime.M(time.Since(startTime).Milliseconds()))

	return nil
}

func (ai *AskIndex) updateIndex() error {
	ctx, cancel := context.WithTimeout(ai.ctx, time.Second*5)
	defer cancel()
	rateLim := make(chan struct{}, queryAskRateLim)
	addrs, err := ai.c.StateListMiners(ctx, nil)
	if err != nil {
		return err
	}

	var lock sync.Mutex
	newIndex := make(map[string]StorageAsk)
	for i, addr := range addrs {
		rateLim <- struct{}{}
		go func(addr string) {
			defer func() { <-rateLim }()
			ctx, cancel := context.WithTimeout(ai.ctx, queryAskTimeout)
			defer cancel()
			pid, err := ai.c.StateMinerPeerID(ctx, addr, nil)
			if err != nil {
				log.Infof("error getting pid of %s: %s", addr, err)
				return
			}
			ask, err := ai.c.ClientQueryAsk(ctx, pid, addr)
			if err != nil {
				return
			}
			lock.Lock()
			newIndex[addr] = StorageAsk{
				Miner:        ask.Ask.Miner,
				Price:        ask.Ask.Price.Uint64(),
				MinPieceSize: ask.Ask.MinPieceSize,
				Timestamp:    ask.Ask.Timestamp,
				Expiry:       ask.Ask.Expiry,
			}
			lock.Unlock()
		}(addr)
		if i%100 == 0 {
			stats.Record(context.Background(), mFullRefreshProgress.M(float64(i)/float64(len(addrs))))
			log.Infof("progress %d/%d", i, len(addrs))
		}
	}
	for i := 0; i < queryAskRateLim; i++ {
		rateLim <- struct{}{}
	}

	median := calculateMedian(newIndex)
	ai.lock.Lock()
	ai.index.LastUpdated = time.Now()
	ai.index.Miners = newIndex
	ai.index.MedianPrice = median
	numAsks := len(ai.index.Miners)
	ai.lock.Unlock()

	stats.Record(context.Background(), mFullRefreshProgress.M(1))
	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "FAIL"))
	stats.Record(ctx, mAskQuery.M(int64(len(addrs)-numAsks)))
	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "OK"))
	stats.Record(ctx, mAskQuery.M(int64(numAsks)))

	return nil
}

func (ai *AskIndex) Close() error {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	if ai.closed {
		return nil
	}
	ai.cancel()
	<-ai.finished
	ai.closed = true
	return nil
}

func calculateMedian(index map[string]StorageAsk) uint64 {
	prices := make([]uint64, 0, len(index))
	for _, v := range index {
		prices = append(prices, v.Price)
	}
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})
	len := len(prices)
	if len < 2 {
		return prices[0]
	}
	if len%2 == 1 {
		return prices[len/2]
	}
	return (prices[len/2-1] + prices[len/2]) / 2
}
