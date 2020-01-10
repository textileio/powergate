package slashing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/chainsync"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
	"go.opencensus.io/stats"
)

var (
	dsBase    = datastore.NewKey("/index/slashing")
	dsHistory = dsBase.ChildString("/index/slashing/history")

	log = logging.Logger("index-slashing")
)

type API interface {
	ChainNotify(context.Context) (<-chan []*types.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
}

type SlashingIndex struct {
	c        API
	ds       datastore.TxnDatastore
	signaler *signaler.Signaler

	lock  sync.Mutex
	index Index

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

type Index struct {
	Tipset types.TipSetKey
	Miners map[string]Info
}

type Info struct {
	History []uint64
}

func New(ds datastore.TxnDatastore, c API) *SlashingIndex {
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	s := &SlashingIndex{
		c:        c,
		ds:       ds,
		signaler: signaler.New(),
		ctx:      ctx,
		index: Index{
			Miners: make(map[string]Info),
		},
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go s.start()

	return s
}

func (s *SlashingIndex) Get() Index {
	s.lock.Lock()
	defer s.lock.Unlock()
	ii := Index{
		Tipset: s.index.Tipset,
		Miners: make(map[string]Info, len(s.index.Miners)),
	}
	for addr, v := range s.index.Miners {
		history := make([]uint64, len(v.History))
		copy(history, v.History)
		ii.Miners[addr] = Info{
			History: history,
		}
	}
	return ii
}
func (s *SlashingIndex) Listen() <-chan struct{} {
	return s.signaler.Listen()
}

func (s *SlashingIndex) Unregister(c chan struct{}) {
	s.signaler.Unregister(c)
}

// ToDo: load from datastore
func (s *SlashingIndex) start() {
	defer close(s.finished)
	n, err := s.c.ChainNotify(s.ctx)
	if err != nil {
		log.Fatalf("error when getting notify channel from lotus: %s", err)
	}
	for {
		select {
		case <-s.ctx.Done():
			log.Info("graceful shutdown of background storage reputation builder")
			return
		case _, ok := <-n:
			if !ok {
				log.Fatalf("lotus notify channel closed: %s", err)
			}
			log.Info("updating slashing index...")
			if err := s.update(); err != nil {
				log.Errorf("error when updating storage reputation: %s", err)
				continue
			}
			log.Info("slashing index updated")
		}
	}
}

func (s *SlashingIndex) update() error {
	base, err := s.lastSyncedTipset()
	if err != nil {
		return err
	}
	path, err := chainsync.GetPathToHead(s.ctx, s.c, base)
	if err != nil {
		return err
	}
	if err = s.updatePath(path); err != nil {
		return fmt.Errorf("error update history: %s", err)
	}
	if err := s.save(); err != nil {
		return err
	}
	s.signaler.Signal()
	return nil
}

func (s *SlashingIndex) lastSyncedTipset() (types.TipSetKey, error) {
	tsk := types.TipSetKey{}
	b, err := s.ds.Get(dsHistory.ChildString("tipset"))
	if err != nil && err != datastore.ErrNotFound {
		return tsk, err
	}
	tsk, err = types.TipSetKeyFromBytes(b)
	if err != nil {
		return tsk, err
	}
	return tsk, nil
}
func (s *SlashingIndex) save() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ds.Put(dsHistory.ChildString("tipset"), s.index.Tipset.Bytes()); err != nil {
		return fmt.Errorf("error saving new height in datastore: %s", err)
	}
	// ToDo: save rest of state in history, may be unnecessary if versioned map does it
	return nil
}

func (s *SlashingIndex) updatePath(path []*types.TipSet) error {
	ctx := context.Background()
	start := time.Now()
	for i := 1; i < len(path); i++ {
		patch, err := historyPatch(s.ctx, s.c, path[i-1], path[i])
		if err != nil {
			return err
		}
		s.lock.Lock()
		// ToDo: should changed to a versioned map history
		state := s.index.Miners
		for addr := range patch {
			info, ok := state[addr]
			if !ok {
				info = Info{}
				state[addr] = info
			}
			if info.History[len(info.History)-1] == patch[addr] {
				continue
			}
			info.History = append(info.History, patch[addr])
		}
		s.lock.Unlock()
		stats.Record(ctx, mRefreshProgress.M(1-float64(i)/float64(len(path))))
	}

	headts := path[len(path)-1]
	stats.Record(ctx, mRefreshDuration.M(int64(time.Since(start).Milliseconds())))
	stats.Record(ctx, mUpdatedHeight.M(int64(headts.Height)))
	stats.Record(ctx, mRefreshProgress.M(1))

	return nil
}

func historyPatch(ctx context.Context, c API, pts *types.TipSet, ts *types.TipSet) (map[string]uint64, error) {
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
				log.Errorf("error when reading state of %s at height %d: %s", addr, pts.Height, err)
				return
			}
			mas, ok := as.State.(map[string]interface{})
			if !ok {
				panic("read state should be a map interface result")
			}

			iSlashedAt, ok := mas["SlashedAt"]
			if !ok {
				log.Warnf("reading state of %s didn't have slashedAt attr", addr)
				return
			}
			fSlashedAt, ok := iSlashedAt.(float64)
			if !ok {
				log.Errorf("casting slashedAt %v from %s at %d failed", iSlashedAt, addr, pts.Height)
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
	return ret, nil
}

func (s *SlashingIndex) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return nil
	}
	s.cancel()
	<-s.finished
	s.closed = true
	return nil
}
