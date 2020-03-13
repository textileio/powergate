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
	cs  ffs.ColdStorage
	hs  ffs.HotStorage
	js  JobStore
	pcs PushConfigStore
	cis CidInfoStore

	work chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

var _ ffs.Scheduler = (*Scheduler)(nil)

// New returns a new instance of Scheduler which uses JobStore as its backing repository for state,
// HotStorage for the hot layer, and ColdStorage for the cold layer.
func New(js JobStore, pcs PushConfigStore, cis CidInfoStore, hs ffs.HotStorage, cs ffs.ColdStorage) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	sch := &Scheduler{
		cs:  cs,
		hs:  hs,
		js:  js,
		pcs: pcs,
		cis: cis,

		work: make(chan struct{}, 1),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go sch.run()
	return sch
}

// PushConfig queues the specified CidConfig to be executed as a new Job. It returns
// the created JobID for further tracking of its state.
func (s *Scheduler) PushConfig(action ffs.PushConfigAction) (ffs.JobID, error) {
	if !action.Config.Cid.Defined() {
		return ffs.EmptyJobID, fmt.Errorf("cid can't be undefined")
	}
	jid := ffs.NewJobID()
	j := ffs.Job{
		ID:         jid,
		InstanceID: action.InstanceID,
		Status:     ffs.Queued,
	}
	if err := s.js.Put(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving push config action in store: %s", err)
	}
	if err := s.pcs.Put(j.ID, action); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving pushed config in store: %s", err)
	}
	select {
	case s.work <- struct{}{}:
	default:
	}
	return jid, nil
}

func (s *Scheduler) GetCidInfo(c cid.Cid) (ffs.CidInfo, error) {
	info, err := s.cis.Get(c)
	if err == ErrNotFound {
		return ffs.CidInfo{}, err
	}
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting CidInfo from store: %s", err)
	}
	return info, nil
}

// GetFromHot returns an io.Reader of the data from the hot layer.
// (TODO: in the future rate-limiting can be applied.)
func (s *Scheduler) GetCidFromHot(ctx context.Context, c cid.Cid) (io.Reader, error) {
	r, err := s.hs.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting %s from hot layer: %s", c, err)
	}
	return r, nil
}

// GetJob the current state of a Job.
func (s *Scheduler) GetJob(jid ffs.JobID) (ffs.Job, error) {
	j, err := s.js.Get(jid)
	if err != nil {
		if err == ErrNotFound {
			return ffs.Job{}, err
		}
		return ffs.Job{}, fmt.Errorf("get Job from store: %s", err)
	}
	return j, nil
}

// Watch returns a channel to listen to Job status changes from a specified
// Api instance. It immediately pushes the current Job state to the channel.
func (s *Scheduler) Watch(iid ffs.InstanceID) <-chan ffs.Job {
	return s.js.Watch(iid)
}

// Unwatch unregisters a subscribing channel created by Watch().
func (s *Scheduler) Unwatch(ch <-chan ffs.Job) {
	s.js.Unwatch(ch)
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
			js, err := s.js.GetByStatus(ffs.Queued)
			if err != nil {
				log.Errorf("getting queued jobs: %s", err)
				continue
			}
			log.Infof("detected %d queued jobs", len(js))
			for _, j := range js {
				log.Infof("executing job %s", j.ID)
				if err := s.mutateJobStatus(j, ffs.InProgress); err != nil {
					log.Errorf("changing job to in-progress: %s", err)
					continue
				}
				info, err := s.execute(s.ctx, j)
				if err != nil {
					j.ErrCause = err.Error()
					if err := s.mutateJobStatus(j, ffs.Failed); err != nil {
						log.Errorf("chanigng job to failed: %s", err)
					}
					log.Errorf("executing job %s: %s", j.ID, err)
					continue
				}
				if err := s.mutateJobStatus(j, ffs.Success); err != nil {
					log.Errorf("changing job to success: %s", err)
					continue
				}
				if err := s.cis.Put(info); err != nil {
					log.Errorf("saving cid info to store: %s", err)
					continue
				}
				log.Infof("job %s executed with final state %d and errcause %q", j.ID, j.Status, j.ErrCause)

			}
		}
	}
}

func (s *Scheduler) execute(ctx context.Context, job ffs.Job) (ffs.CidInfo, error) {
	action, err := s.pcs.Get(job.ID)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting push config action data from store: %s", err)
	}
	hot, err := s.hs.Pin(ctx, action.Config.Cid, action.Config.Hot)
	if err != nil {
		return ffs.CidInfo{}, err
	}

	cold, err := s.cs.Store(ctx, action.Config.Cid, action.WalletAddr, action.Config.Cold)
	if err != nil {
		return ffs.CidInfo{}, err
	}
	return ffs.CidInfo{
		JobID:   job.ID,
		Cid:     action.Config.Cid,
		Hot:     hot,
		Cold:    cold,
		Created: time.Now(),
	}, nil
}

func (s *Scheduler) mutateJobStatus(j ffs.Job, status ffs.JobStatus) error {
	j.Status = status
	if err := s.js.Put(j); err != nil {
		return err
	}
	return nil
}
