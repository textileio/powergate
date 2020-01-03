package miner

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
)

// ToDo: metrics

const (
	fullRefreshThreshold = 100
)

var (
	dsBase = datastore.NewKey("/index/miner")

	apiTimeout      = time.Second * 5
	maxParallelCalc = 10

	log = logging.Logger("index-slashing")
)

type API interface {
	StateListMiners(context.Context, *types.TipSet) ([]string, error)
	StateMinerPower(context.Context, string, *types.TipSet) (types.MinerPower, error)
	ChainNotify(context.Context) (<-chan []*types.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
}

type MinerIndex struct {
	c        API
	ds       datastore.TxnDatastore
	signaler *signaler.Signaler

	info IndexInfo

	lock     sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

type IndexInfo struct {
	LastUpdated uint64
	Power       map[string]*PowerInfo
	// ToDo: geography/region information of miner
}

type PowerInfo struct {
	Power      uint64
	TotalPower uint64
}

func New(ds datastore.TxnDatastore, c API) *MinerIndex {
	ctx, cancel := context.WithCancel(context.Background())
	mi := &MinerIndex{
		c:        c,
		ds:       ds,
		signaler: signaler.New(),
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go mi.start()
	return mi
}

func (mi *MinerIndex) All() *IndexInfo {
	mi.lock.Lock()
	ii := &IndexInfo{
		LastUpdated: mi.info.LastUpdated,
		Power:       make(map[string]*PowerInfo, len(mi.info.Power)),
	}
	for addr := range mi.info.Power {
		v := mi.info.Power[addr]
		ii.Power[addr] = &PowerInfo{
			Power:      v.Power,
			TotalPower: v.TotalPower,
		}
	}
	mi.lock.Unlock()
	return ii
}

func (mi *MinerIndex) start() {
	defer close(mi.finished)
	n, err := mi.c.ChainNotify(mi.ctx)
	if err != nil {
		log.Fatalf("error when getting notify channel from lotus: %s", err)
	}
	for {
		select {
		case <-mi.ctx.Done():
			log.Info("graceful shutdown of background miner index")
			return
		case _, ok := <-n:
			if !ok {
				log.Fatalf("lotus notify channel closed: %s", err) // ToDo: recoverable?
			}
			if err := mi.update(); err != nil {
				log.Errorf("error when updating miner index: %s", err)
				continue
			}
		}
	}
}

func (mi *MinerIndex) Listen() <-chan struct{} {
	return mi.signaler.Listen()
}

func (mi *MinerIndex) Unregister(c chan struct{}) {
	mi.signaler.Unregister(c)
}

func (mi *MinerIndex) update() error {
	lastHeight := uint64(1)
	dsLastHeight := dsBase.ChildString("height")
	b, err := mi.ds.Get(dsLastHeight)
	if err != nil && err != datastore.ErrNotFound {
		return fmt.Errorf("couldn't get stored last height: %s", err)
	}
	if err != datastore.ErrNotFound {
		lastHeight, _ = binary.Uvarint(b)
	}
	log.Debugf("updating from height: %d", lastHeight)

	ts, err := mi.c.ChainHead(mi.ctx) // ToDo: think if this is correct since we could have applys, rollbacks, etc
	if err != nil {
		return fmt.Errorf("error when getting heaviest head from chain: %s", err)
	}

	if ts.Height-lastHeight > fullRefreshThreshold {
		log.Infof("doing full refresh, current %d,  last %d", ts.Height, lastHeight)
		if err := mi.fullRefresh(ts); err != nil {
			return err
		}
	} else {
		log.Infof("doing delta refresh, current %d, last %d", ts.Height, lastHeight)
		if err := mi.deltaRefresh(lastHeight, ts); err != nil {
			return err
		}
	}

	// ToDo: save miner index in datastore
	var buf [32]byte
	n := binary.PutUvarint(buf[:], mi.info.LastUpdated)
	if err := mi.ds.Put(dsLastHeight, buf[:n]); err != nil {
		return fmt.Errorf("error saving new height in datastore: %s", err)
	}
	mi.signaler.Signal()
	return nil
}

func (mi *MinerIndex) deltaRefresh(lastHeight uint64, ts *types.TipSet) error {
	currHeight := ts.Height
	if ts.Height < lastHeight {
		log.Fatal("reversed node or chain rollback, current height %d, last height %d", ts.Height, lastHeight)
	}

	var err error
	path := make([]*types.TipSet, 0, ts.Height-lastHeight)
	for ts.Height != lastHeight+1 {
		path = append(path, ts)
		ts, err = mi.c.ChainGetTipSet(mi.ctx, types.NewTipSetKey(ts.Blocks[0].Parents...))
		if err != nil {
			return err
		}
	}
	var changedMiners map[string]struct{}
	for i := len(path) - 1; i >= 1; i-- {
		chg, err := mi.c.StateChangedActors(mi.ctx, path[i-1].Blocks[0].ParentStateRoot, path[i].Blocks[0].ParentStateRoot)
		if err != nil {
			return err
		}
		for addr := range chg {
			changedMiners[addr] = struct{}{}
		}
	}
	var lock sync.Mutex
	delta := make(map[string]*PowerInfo)
	for addr := range changedMiners {
		r, err := getPower(mi.ctx, mi.c, addr)
		if err != nil {
			log.Errorf("error generating reputation for miner %s: %s", addr, err)
			continue
		}
		lock.Lock()
		delta[addr] = r
		lock.Unlock()
	}

	mi.lock.Lock()
	for addr := range delta {
		mi.info.Power[addr] = delta[addr]
	}
	mi.info.LastUpdated = currHeight
	mi.lock.Unlock()

	return nil
}

func (mi *MinerIndex) fullRefresh(head *types.TipSet) error {
	addrs, err := mi.c.StateListMiners(mi.ctx, nil)
	if err != nil {
		return err
	}
	rateLim := make(chan struct{}, maxParallelCalc)
	var lock sync.Mutex
	newPower := make(map[string]*PowerInfo)
	for _, a := range addrs {
		rateLim <- struct{}{}
		go func(addr string) {
			defer func() { <-rateLim }()
			r, err := getPower(mi.ctx, mi.c, addr)
			if err != nil {
				log.Errorf("error generating reputation for miner %s: %s", addr, err)
				return
			}
			lock.Lock()
			newPower[addr] = r
			lock.Unlock()
		}(a)
	}
	for i := 0; i < maxParallelCalc; i++ {
		rateLim <- struct{}{}
	}

	mi.lock.Lock()
	mi.info.LastUpdated = head.Height
	mi.info.Power = newPower
	mi.lock.Unlock()
	return nil
}

func (mi *MinerIndex) Close() error {
	mi.lock.Lock()
	defer mi.lock.Unlock()
	if mi.closed {
		return nil
	}
	mi.cancel()
	<-mi.finished
	mi.closed = true
	return nil
}

func getPower(ctx context.Context, c API, addr string) (*PowerInfo, error) {
	mp, err := c.StateMinerPower(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return &PowerInfo{
		Power:      mp.MinerPower.Uint64(),
		TotalPower: mp.TotalPower.Uint64(),
	}, nil
}
