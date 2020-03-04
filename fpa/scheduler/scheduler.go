package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/fpa"
)

var (
	log = logging.Logger("scheduler")
)

type Scheduler struct {
	cold  fpa.ColdLayer
	hot   fpa.HotLayer
	store JobStore

	work chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}

	lock   sync.Mutex
	queued []fpa.Job
}

var _ fpa.Scheduler = (*Scheduler)(nil)

func New(store JobStore, hot fpa.HotLayer, cold fpa.ColdLayer) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	sch := &Scheduler{
		store: store,
		hot:   hot,
		cold:  cold,

		work: make(chan struct{}, 1),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go sch.run()
	return sch
}

func (s *Scheduler) Enqueue(c fpa.CidConfig) (fpa.JobID, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	jid := fpa.NewJobID()
	log.Infof("enqueuing %s", jid)
	j := fpa.Job{
		ID:     jid,
		Status: fpa.Queued,
		Config: c,
		CidInfo: fpa.CidInfo{
			ConfigID: c.ID,
			Cid:      c.Cid,
			Created:  time.Now(),
		},
	}
	if err := s.store.Put(j); err != nil {
		return fpa.EmptyJobID, fmt.Errorf("saving enqueued job: %s", err)
	}
	select {
	case s.work <- struct{}{}:
	default:
	}
	return jid, nil
}

func (s *Scheduler) Watch(iid fpa.InstanceID) <-chan fpa.Job {
	return s.store.Watch(iid)
}

func (s *Scheduler) Unwatch(ch <-chan fpa.Job) {
	s.store.Unwatch(ch)
}

func (s *Scheduler) Close() error {
	s.cancel()
	<-s.finished
	return nil
}

func (s *Scheduler) run() {
	defer close(s.finished)
	for {
		select {
		case <-s.ctx.Done():
			log.Infof("terminating scheduler daemon")
			return
		case <-s.work:
			js, err := s.store.GetByStatus(fpa.Queued)
			if err != nil {
				log.Errorf("getting queued jobs: %s", err)
				continue
			}
			log.Infof("detected %d queued jobs", len(js))
			for _, j := range js {
				log.Infof("executing job %s", j.ID)
				if err := s.execute(s.ctx, &j); err != nil {
					log.Errorf("executing job %s: %s", j.ID, err)
					continue
				}
				log.Infof("job %s executed with final state %d and errcause %s", j.ID, j.Status, j.ErrCause)
				if err := s.store.Put(j); err != nil {
					log.Errorf("saving job %s: %s", j.ID, err)
				}
			}
		}
	}
}

func (s *Scheduler) execute(ctx context.Context, job *fpa.Job) error {
	cinfo := fpa.CidInfo{
		ConfigID: job.Config.ID,
		Cid:      job.Config.Cid,
		Created:  time.Now(),
	}
	var err error
	cinfo.Hot, err = s.hot.Pin(ctx, job.Config.Cid)
	if err != nil {
		job.Status = fpa.Failed
		job.ErrCause = err.Error()
		return nil
	}

	cinfo.Cold, err = s.cold.Store(ctx, job.Config.Cid, job.Config.Cold)
	if err != nil {
		job.Status = fpa.Failed
		job.ErrCause = err.Error()
		return nil
	}
	job.CidInfo = cinfo
	job.Status = fpa.Done
	return nil
}
