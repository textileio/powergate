package scheduler

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("ffs-scheduler")
)

// Scheduler receives actions to store a Cid in Hot and Cold layers. These actions are
// created as Jobs which have a lifecycle that can be watched by external actors.
// This Jobs are executed by delegating the work to the Hot and Cold layers configured for
// the scheduler.
type Scheduler struct {
	cold  ffs.ColdStorage
	hot   ffs.HotStorage
	store JobStore

	work chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

var _ ffs.Scheduler = (*Scheduler)(nil)

// New returns a new instance of Scheduler which uses JobStore as its backing repository for state,
// HotStorage for the hot layer, and ColdStorage for the cold layer.
func New(store JobStore, hot ffs.HotStorage, cold ffs.ColdStorage) *Scheduler {
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

// EnqueueCid queues the specified CidConfig to be executed as a new Job. It returns
// the created JobID for further tracking of its state.
func (s *Scheduler) EnqueueCid(c ffs.CidConfig) (ffs.JobID, error) {
	if !c.Cid.Defined() {
		return ffs.EmptyJobID, fmt.Errorf("cid can't be undefined")
	}
	jid := ffs.NewJobID()
	log.Infof("enqueuing %s", jid)
	j := ffs.Job{
		ID:     jid,
		Status: ffs.Queued,
		Config: c,
		CidInfo: ffs.CidInfo{
			ConfigID: c.ID,
			Cid:      c.Cid,
			Created:  time.Now(),
		},
	}
	if err := s.store.Put(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving enqueued job: %s", err)
	}
	select {
	case s.work <- struct{}{}:
	default:
	}
	return jid, nil
}

// GetFromHot returns an io.Reader of the data from the hot layer.
// (TODO: in the future rate-limiting can be applied.)
func (s *Scheduler) GetCidFromHot(ctx context.Context, c cid.Cid) (io.Reader, error) {
	r, err := s.hot.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting %s from hot layer: %s", c, err)
	}
	return r, nil
}

// GetJob the current state of a Job.
func (s *Scheduler) GetJob(jid ffs.JobID) (ffs.Job, error) {
	return s.store.Get(jid)
}

// Watch returns a channel to listen to Job status changes from a specified
// Api instance. It immediately pushes the current Job state to the channel.
func (s *Scheduler) Watch(iid ffs.InstanceID) <-chan ffs.Job {
	return s.store.Watch(iid)
}

// Unwatch unregisters a subscribing channel created by Watch().
func (s *Scheduler) Unwatch(ch <-chan ffs.Job) {
	s.store.Unwatch(ch)
}

// Close terminates the scheduler.
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
			js, err := s.store.GetByStatus(ffs.Queued)
			if err != nil {
				log.Errorf("getting queued jobs: %s", err)
				continue
			}
			log.Infof("detected %d queued jobs", len(js))
			for _, j := range js {
				log.Infof("executing job %s", j.ID)
				j.Status = ffs.InProgress
				if err := s.store.Put(j); err != nil {
					log.Errorf("switching job to InProgress: %s", err)
					continue
				}
				if err := s.execute(s.ctx, &j); err != nil {
					log.Errorf("executing job %s: %s", j.ID, err)
					continue
				}
				log.Infof("job %s executed with final state %d and errcause %q", j.ID, j.Status, j.ErrCause)
				if err := s.store.Put(j); err != nil {
					log.Errorf("saving job %s: %s", j.ID, err)
				}
			}
		}
	}
}

func (s *Scheduler) execute(ctx context.Context, job *ffs.Job) error {
	job.CidInfo = ffs.CidInfo{
		ConfigID: job.Config.ID,
		Cid:      job.Config.Cid,
		Created:  time.Now(),
	}
	var err error
	job.CidInfo.Hot, err = s.hot.Pin(ctx, job.Config.Cid)
	if err != nil {
		job.Status = ffs.Failed
		job.ErrCause = err.Error()
		return nil
	}

	job.CidInfo.Cold, err = s.cold.Store(ctx, job.Config.Cid, job.Config.Cold)
	if err != nil {
		job.Status = ffs.Failed
		job.ErrCause = err.Error()
		return nil
	}
	job.Status = ffs.Done
	return nil
}
