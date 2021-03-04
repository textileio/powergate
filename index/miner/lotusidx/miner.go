package lotusidx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/v2/chainstore"
	"github.com/textileio/powergate/v2/chainsync"
	"github.com/textileio/powergate/v2/index/miner"
	"github.com/textileio/powergate/v2/iplocation"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/signaler"
	txndstr "github.com/textileio/powergate/v2/txndstransform"
)

var (
	minersRefreshInterval = time.Hour * 6
	maxParallelism        = 1
	dsBase                = datastore.NewKey("index")

	log = logging.Logger("index-miner")
)

// P2PHost provides a client to connect to a libp2p peer.
type P2PHost interface {
	Addrs(pid peer.ID) []multiaddr.Multiaddr
	Ping(ctx context.Context, pid peer.ID) bool
	GetAgentVersion(pid peer.ID) string
}

// Index builds and provides information about FC miners.
type Index struct {
	clientBuilder lotus.ClientBuilder
	ds            datastore.TxnDatastore
	store         *chainstore.Store
	h             P2PHost
	lr            iplocation.LocationResolver
	signaler      *signaler.Signaler

	lock  sync.Mutex
	index miner.IndexSnapshot

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed bool
}

// New returns a new MinerIndex. It loads from ds any previous state and starts
// immediately making the index up to date.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, h P2PHost, lr iplocation.LocationResolver, runOnStart, disable bool) (*Index, error) {
	cs := chainsync.New(clientBuilder)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	mi := &Index{
		clientBuilder: clientBuilder,
		ds:            ds,
		store:         store,
		signaler:      signaler.New(),
		h:             h,
		lr:            lr,

		ctx:    ctx,
		cancel: cancel,
	}
	if err := mi.loadFromDS(); err != nil {
		return nil, err
	}

	if !disable {
		mi.startMinerWorker(runOnStart)
		mi.startMetaWorker(runOnStart)
	}
	return mi, nil
}

// Get returns a copy of the current index information.
func (mi *Index) Get() miner.IndexSnapshot {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	ii := miner.IndexSnapshot{
		Meta: miner.MetaIndex{
			Info: make(map[string]miner.Meta, len(mi.index.Meta.Info)),
		},
		OnChain: miner.ChainIndex{
			LastUpdated: mi.index.OnChain.LastUpdated,
			Miners:      make(map[string]miner.OnChainData, len(mi.index.OnChain.Miners)),
		},
	}
	for addr, v := range mi.index.Meta.Info {
		ii.Meta.Info[addr] = v
	}
	for addr, v := range mi.index.OnChain.Miners {
		ii.OnChain.Miners[addr] = v
	}
	return ii
}

// Listen returns a channel signaler to notify when new index information is
// available.
func (mi *Index) Listen() <-chan struct{} {
	return mi.signaler.Listen()
}

// Unregister unregisters a channel signaler from the signaler hub.
func (mi *Index) Unregister(c chan struct{}) {
	mi.signaler.Unregister(c)
}

// Close closes a MinerIndex.
func (mi *Index) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	mi.lock.Lock()
	defer mi.lock.Unlock()
	if mi.closed {
		return nil
	}
	mi.cancel()
	mi.wg.Wait()
	mi.signaler.Close()

	mi.closed = true
	return nil
}

func (mi *Index) startMinerWorker(runOnStart bool) {
	mi.wg.Add(1)
	go func() {
		defer mi.wg.Done()

		startRun := make(chan struct{}, 1)
		if runOnStart {
			startRun <- struct{}{}
		}

		if err := mi.updateOnChainIndex(); err != nil {
			log.Errorf("initial updating miner index: %s", err)
		}
		for {
			select {
			case <-mi.ctx.Done():
				log.Info("graceful shutdown of background miner index")
				return
			case <-time.After(minersRefreshInterval):
				if err := mi.updateOnChainIndex(); err != nil {
					log.Errorf("updating miner index: %s", err)
				}
			case <-startRun:
				if err := mi.updateOnChainIndex(); err != nil {
					log.Errorf("updating miner index on first-run: %s", err)
				}
			}
		}
	}()
}

// loadFromDS loads persisted indexes to memory datastructures. No locks needed
// since its only called from New().
func (mi *Index) loadFromDS() error {
	mi.index = miner.IndexSnapshot{
		Meta:    miner.MetaIndex{Info: make(map[string]miner.Meta)},
		OnChain: miner.ChainIndex{Miners: make(map[string]miner.OnChainData)},
	}
	buf, err := mi.ds.Get(dsKeyMetaIndex)
	if err != nil && err != datastore.ErrNotFound {
		return fmt.Errorf("getting metadata key value: %s", err)
	}
	if err == nil {
		var metaIndex miner.MetaIndex
		if err := json.Unmarshal(buf, &metaIndex); err != nil {
			log.Warnf("wrong json?: %s", string(buf))
			return fmt.Errorf("unmarshaling meta index: %s", err)
		}
		mi.index.Meta = metaIndex
	}

	var chainIndex miner.ChainIndex
	if _, err := mi.store.GetLastCheckpoint(&chainIndex); err != nil {
		return fmt.Errorf("getting last checkpoint: %s", err)
	}
	mi.index.OnChain = chainIndex

	return nil
}
