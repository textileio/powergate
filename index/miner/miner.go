package miner

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/lotus-client/api/apistruct"
	"github.com/textileio/powergate/chainstore"
	"github.com/textileio/powergate/chainsync"
	"github.com/textileio/powergate/iplocation"
	"github.com/textileio/powergate/signaler"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
)

const (
	goroutinesCount = 2
)

var (
	maxParallelism = 10
	dsBase         = datastore.NewKey("index")

	log = logging.Logger("index-miner")
)

// P2PHost provides a client to connect to a libp2p peer.
type P2PHost interface {
	Addrs(pid peer.ID) []multiaddr.Multiaddr
	Ping(ctx context.Context, pid peer.ID) bool
	GetAgentVersion(pid peer.ID) string
}

// Index builds and provides information about FC miners
type Index struct {
	api      *apistruct.FullNodeStruct
	ds       datastore.TxnDatastore
	store    *chainstore.Store
	h        P2PHost
	lr       iplocation.LocationResolver
	signaler *signaler.Signaler

	chMeta      chan struct{}
	metaTicker  *time.Ticker
	minerTicker *time.Ticker
	lock        sync.Mutex
	index       IndexSnapshot

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool
}

// New returns a new MinerIndex. It loads from ds any previous state and starts
// immediately making the index up to date.
func New(ds datastore.TxnDatastore, api *apistruct.FullNodeStruct, h P2PHost, lr iplocation.LocationResolver) (*Index, error) {
	cs := chainsync.New(api)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	mi := &Index{
		api:         api,
		ds:          ds,
		store:       store,
		signaler:    signaler.New(),
		h:           h,
		lr:          lr,
		chMeta:      make(chan struct{}, 1),
		metaTicker:  time.NewTicker(metadataRefreshInterval),
		minerTicker: time.NewTicker(util.AvgBlockTime),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}, 2),
	}
	if err := mi.loadFromDS(); err != nil {
		return nil, err
	}
	go mi.start()
	go mi.metaWorker()
	return mi, nil
}

// Get returns a copy of the current index information
func (mi *Index) Get() IndexSnapshot {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	ii := IndexSnapshot{
		Meta: MetaIndex{
			Online:  mi.index.Meta.Online,
			Offline: mi.index.Meta.Offline,
			Info:    make(map[string]Meta, len(mi.index.Meta.Info)),
		},
		Chain: ChainIndex{
			LastUpdated: mi.index.Chain.LastUpdated,
			Power:       make(map[string]Power, len(mi.index.Chain.Power)),
		},
	}
	for addr, v := range mi.index.Meta.Info {
		ii.Meta.Info[addr] = v
	}
	for addr, v := range mi.index.Chain.Power {
		ii.Chain.Power[addr] = v
	}
	return ii
}

// Listen returns a channel signaler to notify when new index information is
// available.
func (mi *Index) Listen() <-chan struct{} {
	return mi.signaler.Listen()
}

// Unregister unregisters a channel signaler from the signaler hub
func (mi *Index) Unregister(c chan struct{}) {
	mi.signaler.Unregister(c)
}

// Close closes a MinerIndex
func (mi *Index) Close() error {
	log.Info("Closing")
	mi.clsLock.Lock()
	defer mi.clsLock.Unlock()
	if mi.closed {
		return nil
	}
	mi.cancel()
	for i := 0; i < goroutinesCount; i++ {
		<-mi.finished
	}
	close(mi.finished)
	mi.signaler.Close()
	mi.minerTicker.Stop()
	mi.metaTicker.Stop()

	mi.closed = true
	return nil
}

// start is a long running job that keep the index up to date. It separates
// updating tasks in two components. Updating on-chain information whenever
// a new potential tipset is notified by the full node. And a metadata updater
// which do best-efforts to gather/update off-chain information about known
// miners.
func (mi *Index) start() {
	defer func() { mi.finished <- struct{}{} }()

	if err := mi.updateOnChainIndex(); err != nil {
		log.Errorf("error on initial updating miner index: %s", err)
	}
	mi.chMeta <- struct{}{}
	for {
		select {
		case <-mi.ctx.Done():
			log.Info("graceful shutdown of background miner index")
			return
		case <-mi.metaTicker.C:
			select {
			case mi.chMeta <- struct{}{}:
			default:
				log.Info("skipping meta index update since it's busy")
			}
		case <-mi.minerTicker.C:
			if err := mi.updateOnChainIndex(); err != nil {
				log.Errorf("error when updating miner index: %s", err)
				continue
			}
		}
	}
}

// loadFromDS loads persisted indexes to memory datastructures. No locks needed
// since its only called from New().
func (mi *Index) loadFromDS() error {
	mi.index = IndexSnapshot{
		Meta:  MetaIndex{Info: make(map[string]Meta)},
		Chain: ChainIndex{Power: make(map[string]Power)},
	}
	buf, err := mi.ds.Get(dsKeyMetaIndex)
	if err != nil && err != datastore.ErrNotFound {
		return err
	}
	if err == nil {
		var metaIndex MetaIndex
		if err := cbor.DecodeInto(buf, &metaIndex); err != nil {
			return err
		}
		mi.index.Meta = metaIndex
	}

	var chainIndex ChainIndex
	if _, err := mi.store.GetLastCheckpoint(&chainIndex); err != nil {
		return err
	}
	mi.index.Chain = chainIndex

	return nil
}
