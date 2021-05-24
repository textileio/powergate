package runner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/index/ask"
	"github.com/textileio/powergate/v2/index/ask/internal/store"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/signaler"
	"go.opentelemetry.io/otel/metric"
)

var (
	log = logging.Logger("index-ask")
)

// Runner contains cached information about markets.
type Runner struct {
	clientBuilder lotus.ClientBuilder
	store         *store.Store
	signaler      *signaler.Signaler
	config        Config

	lock        sync.Mutex
	index       ask.Index
	orderedAsks []*ask.StorageAsk

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool

	// Metrics
	refreshDuration metric.Int64ValueRecorder
	metricLock      sync.Mutex
	progress        float64
}

// Config contains parameters for index updating.
type Config struct {
	Disable         bool
	QueryAskTimeout time.Duration
	MaxParallel     int
	RefreshInterval time.Duration
	RefreshOnStart  bool
}

// New returns a new ask index runner. It load a persisted ask index, and immediately starts building a new fresh one.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, config Config) (*Runner, error) {
	store := store.New(ds)
	idx, err := store.Get()
	if err != nil {
		return nil, fmt.Errorf("loading from store: %s", err)
	}
	log.Infof("loaded persisted index with %d entries", len(idx.Storage))
	ctx, cancel := context.WithCancel(context.Background())
	ai := &Runner{
		signaler:      signaler.New(),
		clientBuilder: clientBuilder,
		store:         store,
		config:        config,

		index:       idx,
		orderedAsks: generateOrderedAsks(idx.Storage),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	ai.initMetrics()

	go ai.start(config.RefreshOnStart, config.Disable)

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
		if q.PieceSize != 0 && sa.MaxPieceSize < q.PieceSize {
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
func (ai *Runner) start(refreshOnStart bool, disable bool) {
	defer close(ai.finished)
	if refreshOnStart {
		if err := ai.update(); err != nil {
			log.Errorf("updating miners asks: %s", err)
		}
	}
	for {
		select {
		case <-ai.ctx.Done():
			log.Info("graceful shutdown of ask index background job")
			return
		case <-time.After(ai.config.RefreshInterval):
			if disable {
				log.Infof("skipping update since disabled")
				continue
			}
			if err := ai.update(); err != nil {
				log.Errorf("updating miners asks: %s", err)
			}
		}
	}
}

// update triggers a full-scan generates and saves a new fresh index and builds
// views for better querying.
func (ai *Runner) update() error {
	start := time.Now()
	log.Info("updating ask index...")
	defer log.Info("ask index updated")

	client, cls, err := ai.clientBuilder(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	newIndex, cache, err := ai.generateIndex(ai.ctx, client)
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
	ai.signaler.Signal()

	ai.refreshDuration.Record(context.Background(), time.Since(start).Milliseconds())

	return nil
}

// generateIndex returns a fresh index.
func (ai *Runner) generateIndex(ctx context.Context, api *api.FullNodeStruct) (ask.Index, []*ask.StorageAsk, error) {
	addrs, err := api.StateListMiners(ctx, types.EmptyTSK)
	if err != nil {
		return ask.Index{}, nil, err
	}

	rateLim := make(chan struct{}, ai.config.MaxParallel)
	var lock sync.Mutex
	newAsks := make(map[string]ask.StorageAsk)
	for i, addr := range addrs {
		if ctx.Err() != nil {
			break
		}
		rateLim <- struct{}{}
		go func(addr address.Address) {
			defer func() { <-rateLim }()
			sask, ok, err := getMinerStorageAsk(ctx, api, addr, ai.config.QueryAskTimeout)
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
		if i%5000 == 0 {
			log.Infof("progress %d/%d", i, len(addrs))
		}
		ai.metricLock.Lock()
		ai.progress = float64(i) / float64(len(addrs))
		ai.metricLock.Unlock()
	}
	for i := 0; i < ai.config.MaxParallel; i++ {
		rateLim <- struct{}{}
	}
	ai.metricLock.Lock()
	ai.progress = 1
	ai.metricLock.Unlock()

	if ctx.Err() != nil {
		return ask.Index{}, nil, fmt.Errorf("refresh was canceled")
	}

	cache := generateOrderedAsks(newAsks)
	return ask.Index{
		LastUpdated:        time.Now(),
		StorageMedianPrice: calculateMedian(cache),
		Storage:            newAsks,
	}, cache, nil
}

// getMinerStorage ask returns the result of querying the miner for its current Storage Ask.
// If the miner has zero power, it won't be queried returning false.
func getMinerStorageAsk(ctx context.Context, api *api.FullNodeStruct, addr address.Address, askTimeout time.Duration) (ask.StorageAsk, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, askTimeout)
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

	if mi.PeerId == nil {
		return ask.StorageAsk{}, false, nil
	}

	sask, err := api.ClientQueryAsk(ctx, *mi.PeerId, addr)
	if err != nil {
		return ask.StorageAsk{}, false, nil
	}
	return ask.StorageAsk{
		Miner:         sask.Miner.String(),
		Price:         sask.Price.Uint64(),
		VerifiedPrice: sask.VerifiedPrice.Uint64(),
		MinPieceSize:  uint64(sask.MinPieceSize),
		MaxPieceSize:  uint64(sask.MaxPieceSize),
		Timestamp:     int64(sask.Timestamp),
		Expiry:        int64(sask.Expiry),
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
