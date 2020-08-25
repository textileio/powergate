package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/index/ask"
	"github.com/textileio/powergate/index/ask/internal/metrics"
	"github.com/textileio/powergate/index/ask/internal/store"
	"github.com/textileio/powergate/signaler"
	"github.com/textileio/powergate/util"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	qaRatelim         = 50
	qaTimeout         = time.Second * 20
	qaRefreshInterval = 10 * util.AvgBlockTime

	log = logging.Logger("index-ask")
)

// Runner contains cached information about markets.
type Runner struct {
	api      *apistruct.FullNodeStruct
	store    *store.Store
	signaler *signaler.Signaler

	lock        sync.Mutex
	index       ask.Index
	orderedAsks []*ask.StorageAsk

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool
}

// New returns a new ask index runner. It load a persisted ask index, and immediately starts building a new fresh one.
func New(ds datastore.TxnDatastore, api *apistruct.FullNodeStruct) (*Runner, error) {
	if err := metrics.Init(); err != nil {
		return nil, fmt.Errorf("initing metrics: %s", err)
	}
	store := store.New(ds)
	idx, err := store.Get()
	if err != nil {
		return nil, fmt.Errorf("loading from store: %s", err)
	}
	log.Infof("loaded persisted index with %d entries", len(idx.Storage))
	ctx, cancel := context.WithCancel(context.Background())
	ai := &Runner{
		signaler: signaler.New(),
		api:      api,
		store:    store,

		index:       idx,
		orderedAsks: generateOrderedAsks(idx.Storage),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go ai.start()
	return ai, nil
}

// Get returns a copy of the current index data.
func (ai *Runner) Get() ask.Index {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	index := ask.Index{
		LastUpdated:        ai.index.LastUpdated,
		StorageMedianPrice: ai.index.StorageMedianPrice,
		Storage:            make(map[string]ask.StorageAsk, len(ai.index.Storage)),
	}
	for addr, v := range ai.index.Storage {
		index.Storage[addr] = v
	}
	return index
}

// Query executes a query to retrieve active Asks.
func (ai *Runner) Query(q ask.Query) ([]ask.StorageAsk, error) {
	ai.lock.Lock()
	defer ai.lock.Unlock()
	var res []ask.StorageAsk
	offset := q.Offset
	for _, sa := range ai.orderedAsks {
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
func (ai *Runner) Listen() <-chan struct{} {
	return ai.signaler.Listen()
}

// Unregister unregisters a channel signaler from the signaler hub.
func (ai *Runner) Unregister(c chan struct{}) {
	ai.signaler.Unregister(c)
}

// Close closes the AskIndex.
func (ai *Runner) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
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

// start is a long running job that updates asks information in the market.
func (ai *Runner) start() {
	defer close(ai.finished)
	if err := ai.update(); err != nil {
		log.Errorf("updating miners asks: %s", err)
	}
	for {
		select {
		case <-ai.ctx.Done():
			log.Info("graceful shutdown of ask index background job")
			return
		case <-time.After(qaRefreshInterval):
			if err := ai.update(); err != nil {
				log.Errorf("updating miners asks: %s", err)
			}
		}
	}
}

// update triggers a full-scan generates and saves a new fresh index and builds
// views for better querying.
func (ai *Runner) update() error {
	log.Info("updating ask index...")
	defer log.Info("ask index updated")
	startTime := time.Now()
	newIndex, cache, err := generateIndex(ai.ctx, ai.api)
	if err != nil {
		return fmt.Errorf("generating index: %s", err)
	}
	if len(cache) == 0 {
		log.Warnf("ignoring asks save since size is 0")
		return nil
	}

	ai.lock.Lock()
	ai.index = newIndex
	ai.orderedAsks = cache
	ai.lock.Unlock()

	if err := ai.store.Save(newIndex); err != nil {
		return fmt.Errorf("persisting ask index: %s", err)
	}

	stats.Record(context.Background(), metrics.MFullRefreshDuration.M(time.Since(startTime).Milliseconds()))
	ai.signaler.Signal()

	return nil
}

// generateIndex returns a fresh index.
func generateIndex(ctx context.Context, api *apistruct.FullNodeStruct) (ask.Index, []*ask.StorageAsk, error) {
	addrs, err := api.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return ask.Index{}, nil, err
	}

	rateLim := make(chan struct{}, qaRatelim)
	var lock sync.Mutex
	newAsks := make(map[string]ask.StorageAsk)
	for i, addr := range addrs {
		if ctx.Err() != nil {
			break
		}
		rateLim <- struct{}{}
		go func(addr address.Address) {
			defer func() { <-rateLim }()
			sask, ok, err := getMinerStorageAsk(ctx, api, addr)
			if err != nil {
				log.Errorf("getting miner storage ask: %s", err)
				return
			}
			if !ok {
				return
			}
			lock.Lock()
			newAsks[addr.String()] = sask
			lock.Unlock()
		}(addr)
		if i%100 == 0 {
			stats.Record(context.Background(), metrics.MFullRefreshProgress.M(float64(i)/float64(len(addrs))))
			log.Debugf("progress %d/%d", i, len(addrs))
		}
	}
	for i := 0; i < qaRatelim; i++ {
		rateLim <- struct{}{}
	}

	if ctx.Err() != nil {
		return ask.Index{}, nil, fmt.Errorf("refresh was canceled")
	}

	stats.Record(context.Background(), metrics.MFullRefreshProgress.M(1))
	ctx, _ = tag.New(context.Background(), tag.Insert(metrics.TagAskStatus, "FAIL"))
	stats.Record(ctx, metrics.MAskQueryResult.M(int64(len(addrs)-len(newAsks))))
	ctx, _ = tag.New(context.Background(), tag.Insert(metrics.TagAskStatus, "OK"))
	stats.Record(ctx, metrics.MAskQueryResult.M(int64(len(newAsks))))

	cache := generateOrderedAsks(newAsks)
	return ask.Index{
		LastUpdated:        time.Now(),
		StorageMedianPrice: calculateMedian(cache),
		Storage:            newAsks,
	}, cache, nil
}

// getMinerStorage ask returns the result of querying the miner for its current Storage Ask.
// If the miner has zero power, it won't be queried returning false.
func getMinerStorageAsk(ctx context.Context, api *apistruct.FullNodeStruct, addr address.Address) (ask.StorageAsk, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, qaTimeout)
	defer cancel()
	power, err := api.StateMinerPower(ctx, addr, types.EmptyTSK)
	if err != nil {
		return ask.StorageAsk{}, false, fmt.Errorf("getting power %s: %s", addr, err)
	}
	if power.MinerPower.RawBytePower.IsZero() {
		return ask.StorageAsk{}, false, nil
	}
	mi, err := api.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return ask.StorageAsk{}, false, fmt.Errorf("getting miner %s info: %s", addr, err)
	}

	sask, err := api.ClientQueryAsk(ctx, *mi.PeerId, addr)
	if err != nil {
		return ask.StorageAsk{}, false, nil
	}
	return ask.StorageAsk{
		Miner:        sask.Ask.Miner.String(),
		Price:        sask.Ask.Price.Uint64(),
		MinPieceSize: uint64(sask.Ask.MinPieceSize),
		Timestamp:    int64(sask.Ask.Timestamp),
		Expiry:       int64(sask.Ask.Expiry),
	}, true, nil
}

func calculateMedian(orderedAsks []*ask.StorageAsk) uint64 {
	if len(orderedAsks) == 0 {
		return 0
	}
	len := len(orderedAsks)
	if len < 2 {
		return orderedAsks[0].Price
	}
	if len%2 == 1 {
		return orderedAsks[len/2].Price
	}
	return (orderedAsks[len/2-1].Price + orderedAsks[len/2].Price) / 2
}

func generateOrderedAsks(asks map[string]ask.StorageAsk) []*ask.StorageAsk {
	cache := make([]*ask.StorageAsk, 0, len(asks))
	for _, v := range asks {
		cache = append(cache, &v)
	}
	sort.Slice(cache, func(i, j int) bool {
		return cache[i].Price < cache[j].Price
	})
	return cache
}
