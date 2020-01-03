package reputation

import (
	"context"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/reputation/internal/source"
	"sync"
	"time"
)

var (
	updateSourcesInterval = time.Second * 90
	log                   = logging.Logger("reputation")
)

type ReputationModule struct {
	ds      datastore.TxnDatastore
	sources *source.SourceStore

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ds datastore.TxnDatastore) *ReputationModule {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &ReputationModule{
		ds: ds,

		ctx:     ctx,
		cancel:  cancel,
		sources: source.NewStore(ds),
	}

	go rm.updateSources()

	return rm
}

func (rm *ReputationModule) Close() error {
	rm.cancel()
	return nil
}

func (rm *ReputationModule) AddSource(id string, maddr ma.Multiaddr) error {
	return rm.sources.Add(source.Source{Id: id, Maddr: maddr})
}

func (rm *ReputationModule) updateSources() {
	for {
		select {
		case <-rm.ctx.Done():
			log.Debug("terminating background sources update")
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
						log.Error("error refreshing source %s: %s", s.Id, err)
						return
					}
					if err := rm.sources.Update(s); err != nil {
						log.Error("error persisting updated source %s: %s", s.Id, err)
						return
					}
				}(s)
			}
			wg.Wait()
			log.Debug("all sources refreshed successfully")
		}
	}
}
