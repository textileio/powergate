package ask

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/signaler"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	qaRatelim         = 100
	qaTimeout         = time.Second * 10
	qaRefreshInterval = time.Minute
	dsIndex           = datastore.NewKey("index")

	log = logging.Logger("index-ask")
)

// Index contains cached information about markets
type Index struct {
	api      *apistruct.FullNodeStruct
	ds       datastore.TxnDatastore
	signaler *signaler.Signaler

	lock              sync.Mutex
	index             IndexSnapshot
	priceOrderedCache []*StorageAsk

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool
}

// New returnas a new AskIndex. It loads saved information from ds, and immeediatelly
// starts keeping the cache up to date.
func New(ds datastore.TxnDatastore, api *apistruct.FullNodeStruct) (*Index, error) {
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	ai := &Index{
		signaler: signaler.New(),
		api:      api,
		ds:       ds,
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	if err := ai.loadFromStore(); err != nil {
		return nil, err
	}
	go ai.start()
	return ai, nil
}

// Get returns a copy of the current index data
func (ai *Index) Get() IndexSnapshot {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	index := IndexSnapshot{
		LastUpdated:        ai.index.LastUpdated,
		StorageMedianPrice: ai.index.StorageMedianPrice,
		Storage:            make(map[string]StorageAsk, len(ai.index.Storage)),
	}
	for addr, v := range ai.index.Storage {
		index.Storage[addr] = v
	}
	return index
}

// Query executes a query to retrieve active Asks
func (ai *Index) Query(q Query) ([]StorageAsk, error) {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	var res []StorageAsk
	offset := q.Offset
	for _, sa := range ai.priceOrderedCache {
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

// Listen returns a new channel signaler that notifies when the index gets
// updated.
func (ai *Index) Listen() <-chan struct{} {
	return ai.signaler.Listen()
}

// Unregister unregisters a channel signaler from the signaler hub
func (ai *Index) Unregister(c chan struct{}) {
	ai.signaler.Unregister(c)
}

// Close closes the AskIndex
func (ai *Index) Close() error {
	log.Info("Closing")
	ai.clsLock.Lock()
	defer ai.clsLock.Unlock()
	if ai.closed {
		return nil
	}
	ai.cancel()
	<-ai.finished
	ai.signaler.Close()
	ai.closed = true
	return nil
}

// start is a long running job that updates asks information in the market
func (ai *Index) start() {
	defer close(ai.finished)
	if err := ai.update(); err != nil {
		log.Errorf("error when updating miners asks: %s", err)
	}
	for {
		select {
		case <-ai.ctx.Done():
			log.Info("graceful shutdown of ask index background job")
			return
		case <-time.After(qaRefreshInterval):
			if err := ai.update(); err != nil {
				log.Errorf("error when updating miners asks: %s", err)
			}
		}
	}
}

// update triggers a full-scan generates and saves a new fresh index and builds
// views for better querying.
func (ai *Index) update() error {
	log.Info("updating ask index...")
	defer log.Info("ask index updated")
	startTime := time.Now()
	newIndex, err := generateIndex(ai.ctx, ai.api)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(newIndex)
	if err != nil {
		return err
	}
	if err = ai.ds.Put(dsIndex, buf); err != nil {
		return err
	}

	cache := make([]*StorageAsk, 0, len(ai.index.Storage))
	for _, v := range ai.index.Storage {
		cache = append(cache, &v)
	}
	sort.Slice(cache, func(i, j int) bool {
		return cache[i].Price < cache[j].Price
	})

	ai.lock.Lock()
	ai.index = *newIndex
	ai.priceOrderedCache = cache
	ai.lock.Unlock()

	stats.Record(context.Background(), mFullRefreshDuration.M(time.Since(startTime).Milliseconds()))

	return nil
}

// generateIndex returns a fresh index
func generateIndex(ctx context.Context, api *apistruct.FullNodeStruct) (*IndexSnapshot, error) {
	addrs, err := api.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return nil, err
	}

	rateLim := make(chan struct{}, qaRatelim)
	var lock sync.Mutex
	newAsks := make(map[string]StorageAsk)
	for i, addr := range addrs {
		rateLim <- struct{}{}
		go func(addr address.Address) {
			defer func() { <-rateLim }()
			ictx, cancel := context.WithTimeout(ctx, qaTimeout)
			defer cancel()
			mi, err := api.StateMinerInfo(ictx, addr, types.EmptyTSK)
			if err != nil {
				log.Debugf("error getting miner %s info: %s", addr, err)
				return
			}
			ask, err := api.ClientQueryAsk(ictx, mi.PeerId, addr)
			if err != nil {
				log.Debugf("error query-asking miner: %s", err)
				return
			}
			lock.Lock()
			newAsks[addr.String()] = StorageAsk{
				Miner:        ask.Ask.Miner.String(),
				Price:        ask.Ask.Price.Uint64(),
				MinPieceSize: uint64(ask.Ask.MinPieceSize),
				Timestamp:    int64(ask.Ask.Timestamp),
				Expiry:       int64(ask.Ask.Expiry),
			}
			lock.Unlock()
		}(addr)
		if i%100 == 0 {
			stats.Record(context.Background(), mFullRefreshProgress.M(float64(i)/float64(len(addrs))))
			log.Debugf("progress %d/%d", i, len(addrs))
		}
	}
	for i := 0; i < qaRatelim; i++ {
		rateLim <- struct{}{}
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("refresh was canceled")
	default:
	}

	stats.Record(context.Background(), mFullRefreshProgress.M(1))
	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "FAIL"))
	stats.Record(ctx, mAskQueryResult.M(int64(len(addrs)-len(newAsks))))
	ctx, _ = tag.New(context.Background(), tag.Insert(keyAskStatus, "OK"))
	stats.Record(ctx, mAskQueryResult.M(int64(len(newAsks))))

	return &IndexSnapshot{
		LastUpdated:        time.Now(),
		StorageMedianPrice: calculateMedian(newAsks),
		Storage:            newAsks,
	}, nil
}

func calculateMedian(index map[string]StorageAsk) uint64 {
	if len(index) == 0 {
		return 0
	}
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

func (ai *Index) loadFromStore() error {
	buf, err := ai.ds.Get(dsIndex)
	if err != nil {
		if err == datastore.ErrNotFound {
			ai.index = IndexSnapshot{Storage: make(map[string]StorageAsk)}
			return nil
		}
		return err
	}
	if err = json.Unmarshal(buf, &ai.index); err != nil {
		return err
	}
	return nil
}
