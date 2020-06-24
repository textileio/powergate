package faults

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/chainstore"
	"github.com/textileio/powergate/chainsync"
	"github.com/textileio/powergate/signaler"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	"go.opencensus.io/stats"
)

const (
	batchSize = 2000
)

var (
	// hOffset is the # of tipsets from the heaviest chain to
	// consider for index updating; this to reduce sensibility to
	// chain reorgs.
	hOffset = abi.ChainEpoch(20)

	log = logging.Logger("index-faults")
)

// Index builds and provides faults history of miners.
type Index struct {
	api      *apistruct.FullNodeStruct
	store    *chainstore.Store
	signaler *signaler.Signaler

	lock  sync.Mutex
	index IndexSnapshot

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

// New returns a new FaultIndex. It will load previous state from ds, and
// immediately start getting in sync with new on-chain.
func New(ds datastore.TxnDatastore, api *apistruct.FullNodeStruct) (*Index, error) {
	cs := chainsync.New(api)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	initMetrics()
	ctx, cancel := context.WithCancel(context.Background())
	s := &Index{
		api:      api,
		store:    store,
		signaler: signaler.New(),
		index: IndexSnapshot{
			Miners: make(map[string]Faults),
		},
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	if err := s.loadFromDS(); err != nil {
		return nil, err
	}
	go s.start()
	return s, nil
}

// Get returns a copy of the current index information.
func (s *Index) Get() IndexSnapshot {
	s.lock.Lock()
	defer s.lock.Unlock()
	ii := IndexSnapshot{
		TipSetKey: s.index.TipSetKey,
		Miners:    make(map[string]Faults, len(s.index.Miners)),
	}
	for addr, v := range s.index.Miners {
		history := make([]int64, len(v.Epochs))
		copy(history, v.Epochs)
		ii.Miners[addr] = Faults{
			Epochs: history,
		}
	}
	return ii
}

// Listen returns a a signaler channel which signals that index information
// has been updated.
func (s *Index) Listen() <-chan struct{} {
	return s.signaler.Listen()
}

// Unregister frees a channel from the signaler hub.
func (s *Index) Unregister(c chan struct{}) {
	s.signaler.Unregister(c)
}

// Close closes the FaultIndex.
func (s *Index) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	s.cancel()
	<-s.finished
	return nil
}

// start is a long running job that keeps the index up to date with chain updates.
func (s *Index) start() {
	defer close(s.finished)
	if err := s.updateIndex(); err != nil {
		log.Errorf("initial updating faults index: %s", err)
	}
	for {
		select {
		case <-s.ctx.Done():
			log.Info("graceful shutdown of background faults updater")
			return
		case <-time.After(util.AvgBlockTime):
			if err := s.updateIndex(); err != nil {
				log.Errorf("updating faults history: %s", err)
				continue
			}
		}
	}
}

// updateIndex updates current index.
func (s *Index) updateIndex() error {
	log.Info("updating faults index...")

	chainHead, err := s.api.ChainHead(s.ctx)
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	// If the chain is very young, wait a bit for start building the index.
	if chainHead.Height()-hOffset <= 0 {
		return nil
	}

	// Get the tipset hOffset before current head as the target tipset
	// to let the index be built.
	targetTs, err := s.api.ChainGetTipSetByHeight(s.ctx, chainHead.Height()-hOffset, chainHead.Key())
	if err != nil {
		return fmt.Errorf("getting offseted tipset from head: %s", err)
	}

	var index IndexSnapshot
	// Load the last saved index snapshot which is a child of the target TipSet.
	// Might not be s.index because of reorgs.
	indexTs, err := s.store.LoadAndPrune(s.ctx, targetTs.Key(), &index)
	if err != nil {
		return fmt.Errorf("load tipset state: %s", err)
	}
	if index.Miners == nil {
		index.Miners = make(map[string]Faults)
	}

	// Get the tipset path between the indexTs and the targetTs, so to
	// calculate the faults that happened between last saved index and target.
	_, path, err := chainsync.ResolveBase(s.ctx, s.api, indexTs, targetTs.Key())
	if err != nil {
		return fmt.Errorf("resolving base path: %s", err)
	}

	mctx := context.Background()
	start := time.Now()
	for i := 0; i < len(path); i += batchSize {
		j := i + batchSize
		if j > len(path) {
			j = len(path)
		}

		// Get section head and length, and get all faults happened
		// in this section and include it into the updating index.
		sectionLength := abi.ChainEpoch(j - i)
		sectionHeadTs := path[j-1].Key()
		faults, err := s.api.StateAllMinerFaults(s.ctx, sectionLength, sectionHeadTs)
		if err != nil {
			return fmt.Errorf("getting faults from path section: %s", err)
		}
		index.TipSetKey = sectionHeadTs.String()
		for _, f := range faults {
			currFaults := index.Miners[f.Miner.String()]
			currFaults.Epochs = append(currFaults.Epochs, int64(f.Epoch))
			index.Miners[f.Miner.String()] = currFaults
		}

		// Persist partial progress in each finished section-path.
		if err := s.store.Save(s.ctx, types.NewTipSetKey(path[j-1].Cids()...), index); err != nil {
			return fmt.Errorf("saving new index state: %s", err)
		}
		log.Infof("processed from %d to %d", path[i].Height(), path[j-1].Height())

		// Already make available this section-path proccessed information for external
		// use.
		s.lock.Lock()
		s.index = IndexSnapshot{
			TipSetKey: index.TipSetKey,
			Miners:    make(map[string]Faults, len(index.Miners)),
		}
		for k, v := range index.Miners {
			s.index.Miners[k] = v
		}
		s.lock.Unlock()

		// Signal external actors that updated index information is available.
		s.signaler.Signal()

		stats.Record(mctx, mRefreshProgress.M(float64(i)/float64(len(path))))
	}

	stats.Record(mctx, mRefreshDuration.M(time.Since(start).Milliseconds()))
	stats.Record(mctx, mUpdatedHeight.M(int64(targetTs.Height())))
	stats.Record(mctx, mRefreshProgress.M(1))
	log.Info("faults index updated")

	return nil
}

// loadFromDS loads persisted indexes to memory datastructures. No locks needed
// since its only called from New().
func (s *Index) loadFromDS() error {
	var index IndexSnapshot
	if _, err := s.store.GetLastCheckpoint(&index); err != nil {
		return err
	}
	s.index = index
	return nil
}
