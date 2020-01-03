package slashing

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/signaler"
)

// ToDo: metrics for last update/monitoring.

var (
	dsBase      = datastore.NewKey("/index/slashing")
	dsConsensus = dsBase.ChildString("/index/slashing/consensus")
	dsDeal      = datastore.NewKey("/index/slashing/deal")

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

	consensus ConsensusHistory
	// ToDo: deal storage consensus, wait for Lotus implementation

	lock     sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
	closed   bool
}

// ToDo: handle proper forking, recent history, etc?
type ConsensusHistory struct {
	// LastUpdated is the block height of the information
	LastUpdated uint64
	// History has slashed height History of storage miners
	History map[string][]uint64
}

func New(c API, ds datastore.TxnDatastore) *SlashingIndex {
	ctx, cancel := context.WithCancel(context.Background())
	s := &SlashingIndex{
		c:        c,
		ds:       ds,
		signaler: signaler.New(),
		ctx:      ctx,
		consensus: ConsensusHistory{
			History: make(map[string][]uint64),
		},
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go s.start()
	return s
}

func (s *SlashingIndex) ConsensusHistory() ConsensusHistory {
	ch := ConsensusHistory{
		LastUpdated: s.consensus.LastUpdated,
		History:     make(map[string][]uint64, len(s.consensus.History)),
	}
	for k, v := range s.consensus.History {
		ch.History[k] = v
	}
	return ch
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
				log.Fatalf("lotus notify channel closed: %s", err) // ToDo: recoverable?, think about what we should do
			}
			if err := s.update(); err != nil {
				log.Errorf("error when updating storage reputation: %s", err)
				continue
			}
		}
	}
}

func (s *SlashingIndex) update() error {
	lastHeight := uint64(1)
	dsConsensusLastHeightKey := dsConsensus.ChildString("height")
	b, err := s.ds.Get(dsConsensusLastHeightKey)
	if err != nil && err != datastore.ErrNotFound {
		return fmt.Errorf("couldn't get stored last height: %s", err)
	}
	if err != datastore.ErrNotFound {
		lastHeight, _ = binary.Uvarint(b)
	}
	log.Debugf("updating from height: %d", lastHeight)

	if err = s.updateConsensusSlashHistory(lastHeight); err != nil {
		return fmt.Errorf("error getting up to date with slash history from height %d: %s", lastHeight, err)
	}
	// ToDo: save history in datastore
	var buf [32]byte
	n := binary.PutUvarint(buf[:], s.consensus.LastUpdated)
	if err := s.ds.Put(dsConsensusLastHeightKey, buf[:n]); err != nil {
		return fmt.Errorf("error saving new height in datastore: %s", err)
	}
	s.signaler.Signal()

	return nil
}

func (s *SlashingIndex) updateConsensusSlashHistory(lastHeight uint64) error {
	ts, err := s.c.ChainHead(s.ctx) // ToDo: think if this is correct since we could have applys, rollbacks, etc
	if err != nil {
		return fmt.Errorf("error when getting heaviest head from chain: %s", err)
	}
	headHeight := ts.Height
	if ts.Height < lastHeight {
		log.Fatal("reversed node or chain rollback, current height %d, last height %d", ts.Height, lastHeight)
	}

	path := make([]*types.TipSet, 0, ts.Height-lastHeight)
	for ts.Height != lastHeight {
		path = append(path, ts)
		ts, err = s.c.ChainGetTipSet(s.ctx, types.NewTipSetKey(ts.Blocks[0].Parents...))
		if err != nil {
			return err
		}
	}

	newHistory := make(map[string][]uint64, len(s.consensus.History))
	for a := range s.consensus.History {
		newHistory[a] = make([]uint64, len(s.consensus.History[a]))
		for i, v := range s.consensus.History[a] {
			newHistory[a][i] = v
		}
	}

	for i := len(path) - 1; i >= 1; i-- {
		deltaSlashState, err := slashedStateBetweenTipset(s.ctx, s.c, path[i-1], path[i])
		if err != nil {
			return err
		}

		for addr := range deltaSlashState {
			if h, ok := newHistory[addr]; ok {
				if h[len(h)-1] == deltaSlashState[addr] {
					continue
				}
			}
			newHistory[addr] = append(newHistory[addr], deltaSlashState[addr])
		}
	}

	s.lock.Lock()
	s.consensus.History = newHistory
	s.consensus.LastUpdated = headHeight
	s.lock.Unlock()

	return nil
}

func slashedStateBetweenTipset(ctx context.Context, c API, pts *types.TipSet, ts *types.TipSet) (map[string]uint64, error) {
	chg, err := c.StateChangedActors(ctx, pts.Blocks[0].ParentStateRoot, ts.Blocks[0].ParentStateRoot)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]uint64)
	for addr := range chg {
		actor := chg[addr]
		// ToDo: propose new API to get current slashedAt value?
		as, err := c.StateReadState(ctx, &actor, ts)
		if err != nil {
			log.Errorf("error when reading state of %s at height: %d", addr, pts.Height)
			continue
		}
		mas, ok := as.State.(map[string]interface{})
		if !ok {
			panic("read state should be a map interface result")
		}

		iSlashedAt, ok := mas["SlashedAt"]
		if !ok {
			log.Warnf("reading state of %s didn't have slashedAt attr", addr)
			continue
		}
		fSlashedAt, ok := iSlashedAt.(float64)
		if !ok {
			log.Errorf("casting slashedAt %v from %s at %d failed", iSlashedAt, addr, pts.Height)
			continue
		}
		slashedAt := uint64(fSlashedAt)
		if slashedAt != 0 {
			ret[addr] = (slashedAt)
		}
	}
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
