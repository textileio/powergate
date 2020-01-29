package miner

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/filecoin/chainstore"
	"github.com/textileio/filecoin/chainsync"
	"github.com/textileio/filecoin/fchost"
	"github.com/textileio/filecoin/iplocation"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	txndstr "github.com/textileio/filecoin/txndstransform"
	"github.com/textileio/filecoin/util"
)

const (
	goroutinesCount = 2
)

var (
	maxParallelism = 10
	dsBase         = datastore.NewKey("index")

	log = logging.Logger("index-miner")
)

// API provides an abstraction to a Filecoin full-node
type API interface {
	StateListMiners(context.Context, *types.TipSet) ([]string, error)
	StateMinerPower(context.Context, string, *types.TipSet) (types.MinerPower, error)
	ChainHead(context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetTipSetByHeight(context.Context, uint64, *types.TipSet) (*types.TipSet, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
	StateMinerPeerID(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error)
	ChainGetGenesis(context.Context) (*types.TipSet, error)
	ChainGetPath(context.Context, types.TipSetKey, types.TipSetKey) ([]*types.HeadChange, error)
}

// MinerIndex builds and provides information about FC miners
type MinerIndex struct {
	api      API
	ds       datastore.TxnDatastore
	store    *chainstore.Store
	h        *fchost.FilecoinHost
	lr       iplocation.LocationResolver
	signaler *signaler.Signaler

	chMeta chan struct{}
	lock   sync.Mutex
	index  Index

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool
}

// New returns a new MinerIndex. It loads from ds any previous state and starts
// immediately making the index up to date.
func New(ds datastore.TxnDatastore, api API, h *fchost.FilecoinHost, lr iplocation.LocationResolver) (*MinerIndex, error) {
	cs := chainsync.New(api)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	mi := &MinerIndex{
		api:      api,
		ds:       ds,
		store:    store,
		signaler: signaler.New(),
		h:        h,
		lr:       lr,
		chMeta:   make(chan struct{}, 1),
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
func (mi *MinerIndex) Get() Index {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	ii := Index{
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
func (mi *MinerIndex) Listen() <-chan struct{} {
	return mi.signaler.Listen()
}

// Unregister unregisters a channel signaler from the signaler hub
func (mi *MinerIndex) Unregister(c chan struct{}) {
	mi.signaler.Unregister(c)
}

// Close closes a MinerIndex
func (mi *MinerIndex) Close() error {
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

	mi.closed = true
	return nil
}

// start is a long running job that keep the index up to date. It separates
// updating tasks in two components. Updating on-chain information whenever
// a new potential tipset is notified by the full node. And a metadata updater
// which do best-efforts to gather/update off-chain information about known
// miners.
func (mi *MinerIndex) start() {
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
		case <-time.After(metadataRefreshInterval):
			select {
			case mi.chMeta <- struct{}{}:
			default:
				log.Info("skipping meta index update since it's busy")
			}
		case <-time.After(util.AvgBlockTime):
			if err := mi.updateOnChainIndex(); err != nil {
				log.Errorf("error when updating miner index: %s", err)
				continue
			}
		}
	}
}

// loadFromDS loads persisted indexes to memory datastructures. No locks needed
// since its only called from New().
func (mi *MinerIndex) loadFromDS() error {
	mi.index = Index{
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
