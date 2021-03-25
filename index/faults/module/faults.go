package module

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/chainstore"
	"github.com/textileio/powergate/v2/chainsync"
	"github.com/textileio/powergate/v2/index/faults"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/signaler"
	txndstr "github.com/textileio/powergate/v2/txndstransform"
)

const (
	batchSize = 2000
)

var (
	// hOffset is the # of tipsets from the heaviest chain to
	// consider for index updating; this to reduce sensibility to
	// chain reorgs.
	hOffset = abi.ChainEpoch(20)

	// updateInterval is the interval duration where the fault index
	// will be updated.
	// Note: currently the Fault index is disabled until Lotus re-enables
	// StateAllMinersFaults() API.
	updateInterval = time.Hour * 24 * 365

	log = logging.Logger("index-faults")
)

// Index builds and provides faults history of miners.
type Index struct {
	clientBuilder lotus.ClientBuilder
	store         *chainstore.Store
	signaler      *signaler.Signaler

	lock  sync.Mutex
	index faults.IndexSnapshot

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

// New returns a new FaultIndex. It will load previous state from ds, and
// immediately start getting in sync with new on-chain.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, disable bool) (*Index, error) {
	cs := chainsync.New(clientBuilder)
	store, err := chainstore.New(txndstr.Wrap(ds, "chainstore"), cs)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Index{
		clientBuilder: clientBuilder,
		store:         store,
		signaler:      signaler.New(),
		index: faults.IndexSnapshot{
			Miners: make(map[string]faults.Faults),
		},
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	if err := s.loadFromDS(); err != nil {
		return nil, err
	}

	go s.start(disable)

	return s, nil
}

// Get returns a copy of the current index information.
func (s *Index) Get() faults.IndexSnapshot {
	s.lock.Lock()
	defer s.lock.Unlock()
	ii := faults.IndexSnapshot{
		TipSetKey: s.index.TipSetKey,
		Miners:    make(map[string]faults.Faults, len(s.index.Miners)),
	}
	for addr, v := range s.index.Miners {
		history := make([]int64, len(v.Epochs))
		copy(history, v.Epochs)
		ii.Miners[addr] = faults.Faults{
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
func (s *Index) start(disabled bool) {
	defer close(s.finished)
	for {
		select {
		case <-s.ctx.Done():
			log.Info("graceful shutdown of background faults updater")
			return
		case <-time.After(updateInterval):
			if disabled {
				log.Infof("skipping updating since it's disabled")
				continue
			}
			if err := s.updateIndex(); err != nil {
				log.Errorf("updating faults history: %s", err)
				continue
			}
		}
	}
}

// updateIndex updates current index.
func (s *Index) updateIndex() error {
	client, cls, err := s.clientBuilder(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	log.Info("updating faults index...")

	chainHead, err := client.ChainHead(s.ctx)
	if err != nil {
		return fmt.Errorf("getting chain head: %s", err)
	}
	// If the chain is very young, wait a bit for start building the index.
	if chainHead.Height()-hOffset <= 0 {
		return nil
	}

	// Get the tipset hOffset before current head as the target tipset
	// to let the index be built.
	targetTs, err := client.ChainGetTipSetByHeight(s.ctx, chainHead.Height()-hOffset, chainHead.Key())
	if err != nil {
		return fmt.Errorf("getting offseted tipset from head: %s", err)
	}

	var index faults.IndexSnapshot
	// Load the last saved index snapshot which is a child of the target TipSet.
	// Might not be s.index because of reorgs.
	indexTs, err := s.store.LoadAndPrune(s.ctx, targetTs.Key(), &index)
	if err != nil {
		return fmt.Errorf("load tipset state: %s", err)
	}
	if index.Miners == nil {
		index.Miners = make(map[string]faults.Faults)
	}

	// Get the tipset path between the indexTs and the targetTs, so to
	// calculate the faults that happened between last saved index and target.
	_, path, err := chainsync.ResolveBase(s.ctx, client, indexTs, targetTs.Key())
	if err != nil {
		return fmt.Errorf("resolving base path: %s", err)
	}

	for i := 0; i < len(path); i += batchSize {
		j := i + batchSize
		if j > len(path) {
			j = len(path)
		}

		// Get section head and length, and get all faults happened
		// in this section and include it into the updating index.
		sectionLength := abi.ChainEpoch(j - i)
		sectionHeadTs := path[j-1].Key()
		fs, err := client.StateAllMinerFaults(s.ctx, sectionLength, sectionHeadTs)
		if err != nil {
			return fmt.Errorf("getting faults from path section: %s", err)
		}
		index.TipSetKey = sectionHeadTs.String()
		for _, f := range fs {
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
		s.index = faults.IndexSnapshot{
			TipSetKey: index.TipSetKey,
			Miners:    make(map[string]faults.Faults, len(index.Miners)),
		}
		for k, v := range index.Miners {
			s.index.Miners[k] = v
		}
		s.lock.Unlock()

		// Signal external actors that updated index information is available.
		s.signaler.Signal()
	}

	log.Info("faults index updated")

	return nil
}

// loadFromDS loads persisted indexes to memory datastructures. No locks needed
// since its only called from New().
func (s *Index) loadFromDS() error {
	var index faults.IndexSnapshot
	if _, err := s.store.GetLastCheckpoint(&index); err != nil {
		return err
	}
	s.index = index
	return nil
}
