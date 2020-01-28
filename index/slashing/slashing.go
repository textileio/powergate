package slashing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/chainstore"
	"github.com/textileio/filecoin/chainsync"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	txndstr "github.com/textileio/filecoin/txndstransform"
	"go.opencensus.io/stats"
)

const (
	batchSize = 20
)

var (
	log = logging.Logger("index-slashing")
)

// API provides an abstraction to a Filecoin full-node
type API interface {
	ChainNotify(context.Context) (<-chan []*types.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
	ChainGetPath(context.Context, types.TipSetKey, types.TipSetKey) ([]*types.HeadChange, error)
	ChainGetGenesis(context.Context) (*types.TipSet, error)
}

// SlashingIndex builds and provides slashing history of miners
type SlashingIndex struct {
	api      API
	store    *chainstore.Store
	signaler *signaler.Signaler

	lock  sync.Mutex
	index Index

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	clsLock  sync.Mutex
	closed   bool
}

// New returns a new SlashingIndex. It will load previous state from ds, and
// immediatelly start getting in sync with new on-chain.
func New(ds datastore.TxnDatastore, api API) (*SlashingIndex, error) {
	cs := chainsync.New(api)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	s := &SlashingIndex{
		api:      api,
		store:    store,
		signaler: signaler.New(),
		index: Index{
			Miners: make(map[string]Slashes),
		},
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go s.start()
	return s, nil
}

// Get returns a copy of the current index information
func (s *SlashingIndex) Get() Index {
	s.lock.Lock()
	defer s.lock.Unlock()
	ii := Index{
		TipSetKey: s.index.TipSetKey,
		Miners:    make(map[string]Slashes, len(s.index.Miners)),
	}
	for addr, v := range s.index.Miners {
		history := make([]uint64, len(v.Epochs))
		copy(history, v.Epochs)
		ii.Miners[addr] = Slashes{
			Epochs: history,
		}
	}
	return ii
}

// Listen returns a a signaler channel which signals that index information
// has been updated.
func (s *SlashingIndex) Listen() <-chan struct{} {
	return s.signaler.Listen()
}

// Unregister frees a channel from the signaler hub
func (s *SlashingIndex) Unregister(c chan struct{}) {
	s.signaler.Unregister(c)
}

// Close closes the SlashindIndex
func (s *SlashingIndex) Close() error {
	log.Info("Closing")
	s.clsLock.Lock()
	defer s.clsLock.Unlock()
	if s.closed {
		return nil
	}
	s.cancel()
	<-s.finished
	s.closed = true
	return nil
}

// start is a long running job that keeps the index up to date with chain updates
func (s *SlashingIndex) start() {
	defer close(s.finished)
	n, err := s.api.ChainNotify(s.ctx)
	if err != nil {
		log.Fatalf("error when getting notify channel from lotus: %s", err)
	}
	for {
		select {
		case <-s.ctx.Done():
			log.Info("graceful shutdown of background slashing updater")
			return
		case hcs, ok := <-n:
			if !ok {
				log.Error("lotus notify channel closed")
				return
			}
			log.Info("updating slashing index...")
			currHeadTsk := types.NewTipSetKey(hcs[len(hcs)-1].Val.Cids...)
			if err := s.updateIndex(currHeadTsk); err != nil {
				log.Errorf("error when updating slashing history: %s", err)
				continue
			}
			log.Info("slashing index updated")
		}
	}
}

// updateIndex updates current index with a new discovered chain head.
func (s *SlashingIndex) updateIndex(new types.TipSetKey) error {
	var index Index
	ts, err := s.store.LoadAndPrune(s.ctx, new, &index)
	if err != nil {
		return err
	}
	if index.Miners == nil {
		index.Miners = make(map[string]Slashes)
	}
	_, path, err := chainsync.ResolveBase(s.ctx, s.api, ts, new)
	if err != nil {
		return err
	}
	mctx := context.Background()
	start := time.Now()
	for i := 0; i < len(path); i += batchSize {
		j := i + batchSize
		if j > len(path) {
			j = len(path)
		}
		if err := updateFromPath(s.ctx, s.api, &index, path[i:j]); err != nil {
			return err
		}
		if err := s.store.Save(s.ctx, types.NewTipSetKey(path[j-1].Cids...), index); err != nil {
			return err
		}
		stats.Record(mctx, mRefreshProgress.M(float64(i)/float64(len(path))))
	}

	headts := path[len(path)-1]
	stats.Record(mctx, mRefreshDuration.M(int64(time.Since(start).Milliseconds())))
	stats.Record(mctx, mUpdatedHeight.M(int64(headts.Height)))
	stats.Record(mctx, mRefreshProgress.M(1))

	s.lock.Lock()
	s.index = index
	s.lock.Unlock()

	s.signaler.Signal()
	return nil
}

// updateFromPath updates a saved index state walking a chain path. The path
// usually should be the next epoch from index up to the current head TipSet.
func updateFromPath(ctx context.Context, api API, index *Index, path []*types.TipSet) error {
	for i := 1; i < len(path); i++ {
		patch, err := epochPatch(ctx, api, path[i-1], path[i])
		if err != nil {
			return err
		}
		for addr := range patch {
			info, ok := index.Miners[addr]
			if !ok {
				info = Slashes{}
				index.Miners[addr] = info
			}
			if len(info.Epochs) > 0 && info.Epochs[len(info.Epochs)-1] == patch[addr] {
				continue
			}
			info.Epochs = append(info.Epochs, patch[addr])
		}
	}
	index.TipSetKey = types.NewTipSetKey(path[len(path)-1].Cids...).String()

	return nil
}

// epochPatch returns a map of slashedAt values for miners that changed between
// two consecutive epochs.
func epochPatch(ctx context.Context, c API, pts *types.TipSet, ts *types.TipSet) (map[string]uint64, error) {
	if !areConsecutiveEpochs(pts, ts) {
		return nil, fmt.Errorf("epoch patch can only be called between parent-child tipsets")
	}

	chg, err := c.StateChangedActors(ctx, pts.Blocks[0].ParentStateRoot, ts.Blocks[0].ParentStateRoot)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]uint64)
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(chg))
	for addr := range chg {
		go func(addr string) {
			defer wg.Done()
			actor := chg[addr]
			as, err := c.StateReadState(ctx, &actor, ts)
			if err != nil {
				log.Debugf("error when reading state of %s at height %d: %s", addr, pts.Height, err)
				return
			}
			mas, ok := as.State.(map[string]interface{})
			if !ok {
				log.Debugf("read state should be a map interface result: %#v", as.State)
			}
			iSlashedAt, ok := mas["SlashedAt"]
			if !ok {
				log.Debugf("reading state of %s didn't have slashedAt attr", addr)
				return
			}
			fSlashedAt, ok := iSlashedAt.(float64)
			if !ok {
				log.Debugf("casting slashedAt %v from %s at %d failed", iSlashedAt, addr, pts.Height)
				return
			}
			slashedAt := uint64(fSlashedAt)
			if slashedAt != 0 {
				lock.Lock()
				ret[addr] = (slashedAt)
				lock.Unlock()
			}
		}(addr)
	}
	wg.Wait()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("cancelled by context")
	default:
	}

	return ret, nil
}

func areConsecutiveEpochs(pts, ts *types.TipSet) bool {
	if pts.Height >= ts.Height {
		return false
	}
	cidsP := pts.Cids
	cidsC := ts.Blocks[0].Parents
	if len(cidsP) != len(cidsC) {
		return false
	}
	for i, c := range cidsP {
		if !c.Equals(cidsC[i]) {
			return false
		}
	}
	return true
}
