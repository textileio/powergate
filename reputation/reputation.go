package reputation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/index/ask"
	"github.com/textileio/powergate/index/faults"
	"github.com/textileio/powergate/index/miner"
	"github.com/textileio/powergate/reputation/internal/source"
)

var (
	updateSourcesInterval = time.Second * 90
	log                   = logging.Logger("reputation")
)

// Module consolidates different sources of information to create a
// reputation rank of FC miners.
type Module struct {
	ds      datastore.TxnDatastore
	sources *source.Store

	mi miner.Module
	fi faults.Module
	ai ask.Module

	lockIndex sync.Mutex
	mIndex    miner.IndexSnapshot
	fIndex    faults.IndexSnapshot
	aIndex    ask.Index

	lockScores sync.Mutex
	rebuild    chan struct{}
	scores     []MinerScore

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

// MinerScore contains a score for a miner.
type MinerScore struct {
	Addr  string
	Score int
}

// New returns a new reputation Module.
func New(ds datastore.TxnDatastore, mi miner.Module, fi faults.Module, ai ask.Module) *Module {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &Module{
		ds: ds,
		mi: mi,
		fi: fi,
		ai: ai,

		mIndex: mi.Get(),
		fIndex: fi.Get(),
		aIndex: ai.Get(),

		rebuild:  make(chan struct{}, 1),
		ctx:      ctx,
		cancel:   cancel,
		sources:  source.NewStore(ds),
		finished: make(chan struct{}),
	}

	go rm.updateSources()
	go rm.indexBuilder()
	go rm.subscribeIndexes()

	return rm
}

// AddSource adds a new external Source to be considered for reputation generation.
func (rm *Module) AddSource(id string, maddr ma.Multiaddr) error {
	return rm.sources.Add(source.Source{ID: id, Maddr: maddr})
}

// QueryMiners makes a filtered query on the scored-sorted miner list.
// Empty filter slices represent no-filters applied.
func (rm *Module) QueryMiners(excludedMiners []string, countryCodes []string, trustedMiners []string) ([]MinerScore, error) {
	rm.lockScores.Lock()
	defer rm.lockScores.Unlock()

	var mr []MinerScore
	pmr := make(map[string]struct{})
	for _, tm := range trustedMiners {
		for _, m := range rm.scores {
			if m.Addr == tm {
				pmr[tm] = struct{}{}
				mr = append(mr, m)
				break
			}
		}
	}
	for _, m := range rm.scores {
		if _, ok := pmr[m.Addr]; ok {
			continue
		}
		skip := false
		for _, mb := range excludedMiners {
			if mb == m.Addr {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if len(countryCodes) != 0 {
			minerMeta, ok := rm.mIndex.Meta.Info[m.Addr]
			if !ok {
				continue
			}
			skip := true
			for _, country := range countryCodes {
				if country == minerMeta.Location.Country {
					skip = false
					break
				}
			}
			if skip {
				continue
			}
		}
		mr = append(mr, m)
	}
	return mr, nil
}

// GetTopMiners gets the top n miners with best score.
func (rm *Module) GetTopMiners(n int) ([]MinerScore, error) {
	if n < 1 {
		return nil, fmt.Errorf("the number of miners should be greater than zero")
	}
	rm.lockScores.Lock()
	if n > len(rm.scores) {
		n = len(rm.scores)
	}
	mr := make([]MinerScore, 0, n)
	for i := 0; i < n; i++ {
		mr = append(mr, rm.scores[i])
	}
	rm.lockScores.Unlock()
	return mr, nil
}

// Close closes the reputation Module.
func (rm *Module) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	rm.cancel()
	<-rm.finished
	return nil
}

// subscribeIndexes listen to all sources changes to trigger score regeneration.
func (rm *Module) subscribeIndexes() {
	defer close(rm.finished)
	subMi := rm.mi.Listen()
	subSi := rm.fi.Listen()
	subAi := rm.ai.Listen()

	rm.rebuild <- struct{}{}
	for {
		select {
		case <-rm.ctx.Done():
			log.Info("terminating background index update")
			return
		case <-subMi:
			rm.lockIndex.Lock()
			rm.mIndex = rm.mi.Get()
		case <-subSi:
			rm.lockIndex.Lock()
			rm.fIndex = rm.fi.Get()
		case <-subAi:
			rm.lockIndex.Lock()
			rm.aIndex = rm.ai.Get()
		}
		rm.lockIndex.Unlock()
		select {
		case rm.rebuild <- struct{}{}:
		default:
		}
	}
}

// indexBuilder regenerates score information from all known sources.
func (rm *Module) indexBuilder() {
	for range rm.rebuild {
		log.Info("rebuilding index")
		start := time.Now()

		sources, err := rm.sources.GetAll()
		if err != nil {
			log.Errorf("getting sources: %s", err)
			return
		}
		rm.lockIndex.Lock()
		minerIndex := rm.mIndex
		faultsIndex := rm.fIndex
		askIndex := rm.aIndex
		rm.lockIndex.Unlock()

		scores := make([]MinerScore, 0, len(askIndex.Storage))
		for addr := range askIndex.Storage {
			score := calculateScore(addr, minerIndex, faultsIndex, askIndex, sources)
			scores = append(scores, score)
		}
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].Score > scores[j].Score
		})

		rm.lockScores.Lock()
		rm.scores = scores
		rm.lockScores.Unlock()

		log.Infof("built scores for %d miners in %dms", len(scores), time.Since(start).Milliseconds())
	}
}

// calculateScore calculates the score for a miner.
func calculateScore(addr string, mi miner.IndexSnapshot, si faults.IndexSnapshot, ai ask.Index, ss []source.Source) MinerScore {
	miner := mi.OnChain.Miners[addr]
	powerScore := miner.RelativePower

	var faultsScore float64
	if faults, ok := si.Miners[addr]; ok {
		faultsScore = 1 / math.Pow(2, float64(len(faults.Epochs)))
	}

	var externalScore float64
	for _, s := range ss {
		score, exist := s.Scores[addr]
		if !exist {
			continue
		}
		externalScore = s.Weight * float64(score)
	}

	var askScore float64
	if a, ok := ai.Storage[addr]; ok && a.Price < ai.StorageMedianPrice {
		askScore = 1
	}

	score := 50*faultsScore + 20*powerScore + 20*externalScore + 10*askScore + 1000*float64(miner.ActiveDeals)
	return MinerScore{
		Addr:  addr,
		Score: int(score),
	}
}

func (rm *Module) updateSources() {
	for {
		select {
		case <-rm.ctx.Done():
			log.Info("terminating background sources update")
			return
		case <-time.After(updateSourcesInterval):
			sources, err := rm.sources.GetAll()
			if err != nil {
				log.Errorf("error getting all sources from store: %s", err)
				continue
			}
			var wg sync.WaitGroup
			wg.Add(len(sources))
			for _, s := range sources {
				go func(s source.Source) {
					defer wg.Done()
					if err := s.Refresh(rm.ctx); err != nil {
						log.Error("error refreshing source %s: %s", s.ID, err)
						return
					}
					if err := rm.sources.Update(s); err != nil {
						log.Error("error persisting updated source %s: %s", s.ID, err)
						return
					}
				}(s)
			}
			wg.Wait()
			log.Debug("all sources refreshed successfully")
		}
	}
}
