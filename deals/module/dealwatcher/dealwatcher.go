package dealwatcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/lotus"
	"go.opentelemetry.io/otel/metric"
)

var (
	log = logger.Logger("deals-watcher")

	ErrNotFound           = errors.New("subscription not found")
	ErrActiveSubscription = errors.New("active subscription")
)

type DealWatcher struct {
	cb lotus.ClientBuilder

	lock sync.Mutex
	subs map[cid.Cid][]chan<- struct{}

	closeLock     sync.Mutex
	closeCtx      context.Context
	closeCancel   context.CancelFunc
	closeFinished chan struct{}
	closed        bool

	// Metrics
	metricDealUpdates            metric.Int64Counter
	metricDealUpdatesChanFailure metric.Int64Counter
}

func New(cb lotus.ClientBuilder) (*DealWatcher, error) {
	ctx, cls := context.WithCancel(context.Background())
	dw := &DealWatcher{
		cb:            cb,
		subs:          make(map[cid.Cid][]chan<- struct{}),
		closeCtx:      ctx,
		closeCancel:   cls,
		closeFinished: make(chan struct{}),
	}

	if err := dw.startDaemon(); err != nil {
		return nil, fmt.Errorf("starting daemon: %s", err)
	}

	dw.initMetrics()

	return dw, nil
}

func (dw *DealWatcher) Subscribe(ch chan<- struct{}, proposalCid cid.Cid) error {
	dw.lock.Lock()
	defer dw.lock.Unlock()

	for _, ich := range dw.subs[proposalCid] {
		if ch == ich {
			return ErrActiveSubscription
		}
	}

	dw.subs[proposalCid] = append(dw.subs[proposalCid], ch)

	log.Infof("subscriber registered")
	return nil
}

func (dw *DealWatcher) Unsubscribe(ch chan<- struct{}, proposalCid cid.Cid) error {
	dw.lock.Lock()
	defer dw.lock.Unlock()

	subs, ok := dw.subs[proposalCid]
	if !ok {
		return ErrNotFound
	}
	idx := -1
	for i := range subs {
		if subs[i] == ch {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrNotFound
	}
	if len(subs) == 1 {
		delete(dw.subs, proposalCid)
		return nil
	}
	subs[idx] = subs[len(subs)-1]
	subs = subs[:len(subs)-1]
	dw.subs[proposalCid] = subs

	return nil
}

func (dw *DealWatcher) Close() error {
	dw.closeLock.Lock()
	defer dw.closeLock.Unlock()

	if dw.closed {
		return nil
	}
	dw.closed = true

	dw.closeCancel()
	<-dw.closeFinished

	return nil
}

func (dw *DealWatcher) startDaemon() error {
	createUpdateChan := func() (<-chan api.DealInfo, func(), error) {
		c, cls, err := dw.cb(dw.closeCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating lotus client: %s", err)
		}

		updates, err := c.ClientGetDealUpdates(dw.closeCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating lotus deal updates channel: %s", err)
		}
		return updates, cls, nil
	}

	go func() {
		updates, cls, err := createUpdateChan()
		if err != nil {
			log.Errorf("creating initial updates chan: %s", err)
			return
		}
		log.Infof("deal watcher created")
		defer close(dw.closeFinished)
		defer func() { cls() }()

		for {
			select {
			case <-dw.closeCtx.Done():
				return
			case di, ok := <-updates:
				if !ok {
					if dw.closeCtx.Err() != nil {
						return
					}

					dw.metricDealUpdatesChanFailure.Add(dw.closeCtx, 1)
					log.Warnf("updates channel closed unexpectedly")

					cls() // Formally closed broken chan.
					for {
						updates, cls, err = createUpdateChan()
						if err != nil {
							log.Warnf("reconstructing updates channel: %s", err)
							time.Sleep(time.Second * 30)
							continue
						}
						break
					}
				}

				dw.lock.Lock()

				subs, ok := dw.subs[di.ProposalCid]
				if !ok {
					dw.lock.Unlock()

					dw.metricDealUpdates.Add(dw.closeCtx, 1, attrDealUntracked)
					continue
				}
				dw.metricDealUpdates.Add(dw.closeCtx, 1, attrDealTracked)
				for _, s := range subs {
					select {
					case s <- struct{}{}:
					default:
						log.Warn("skipping slow receiver")
					}
				}
				dw.lock.Unlock()
			}
		}
	}()

	return nil
}
