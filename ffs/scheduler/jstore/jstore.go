package jstore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sched-jstore")
)

// Store is a Datastore implementation of JobStore, which saves
// state of scheduler Jobs.
type Store struct {
	lock       sync.Mutex
	ds         datastore.Datastore
	watchers   []watcher
	inProgress map[cid.Cid]struct{}
}

var _ scheduler.JobStore = (*Store)(nil)

type watcher struct {
	iid ffs.APIID
	C   chan ffs.Job
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds:         ds,
		inProgress: make(map[cid.Cid]struct{}),
	}
}

// Finalize sets a Job status to a final state, i.e. Success or Failed,
// with a list of Deal errors ocurred during job execution.
func (s *Store) Finalize(jid ffs.JobID, st ffs.JobStatus, errors []ffs.DealError) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	j, err := s.get(jid)
	if err != nil {
		return err
	}
	if st != ffs.Success && st != ffs.Failed {
		return fmt.Errorf("new state should be final")
	}
	j.Status = st
	j.DealErrors = errors
	if err := s.put(j); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// Dequeue dequeues a Job which doesn't have have another in-progress Job
// for the same Cid. Saying it differently, it's safe to execute. The returned
// job Status is automatically changed to Queued. If no jobs are available to dequeue
// it returns a nil *ffs.Job and no-error.
func (s *Store) Dequeue() (*ffs.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{Prefix: ""}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing dequeue query result: %s", err)
		}
	}()
	for r := range res.Next() {
		var j ffs.Job
		if err := json.Unmarshal(r.Value, &j); err != nil {
			return nil, fmt.Errorf("unmarshalling job: %s", err)
		}
		_, ok := s.inProgress[j.Cid]
		if j.Status == ffs.Queued && !ok {
			j.Status = ffs.InProgress
			if err := s.put(j); err != nil {
				return nil, err
			}
			return &j, nil
		}
	}
	return nil, nil
}

// Enqueue queues a new Job. If other Job for the same Cid is in Queued status,
// it will be automatically marked as Canceled.
func (s *Store) Enqueue(j ffs.Job) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.cancelQueued(j.Cid); err != nil {
		return fmt.Errorf("canceling queued jobs: %s", err)
	}
	j.Status = ffs.Queued
	if err := s.put(j); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}
	return nil
}

func (s *Store) cancelQueued(c cid.Cid) error {
	q := query.Query{Prefix: ""}
	res, err := s.ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing getbystatus query result: %s", err)
		}
	}()
	for r := range res.Next() {
		var j ffs.Job
		if err := json.Unmarshal(r.Value, &j); err != nil {
			return fmt.Errorf("unmarshalling job: %s", err)
		}
		if j.Status == ffs.Queued && j.Cid == c {
			j.Status = ffs.Canceled
			if err := s.put(j); err != nil {
				return fmt.Errorf("canceling queued job: %s", err)
			}
		}
	}
	return nil
}

// Get returns the current state of Job. If doesn't exist, returns
// ErrNotFound.
func (s *Store) Get(jid ffs.JobID) (ffs.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.get(jid)
}

// Watch subscribes to Job changes from a specified Api instance.
func (s *Store) Watch(ctx context.Context, c chan<- ffs.Job, iid ffs.APIID) error {
	s.lock.Lock()
	ic := make(chan ffs.Job, 1)
	s.watchers = append(s.watchers, watcher{iid: iid, C: ic})
	s.lock.Unlock()

	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			stop = true
		case l, ok := <-ic:
			if !ok {
				return fmt.Errorf("jobstore was closed with a listening client")
			}
			c <- l
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	for i := range s.watchers {
		if s.watchers[i].C == ic {
			s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
			break
		}
	}
	return nil
}

// Close closes the Store, unregistering any subscribed watchers.
func (s *Store) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i := range s.watchers {
		close(s.watchers[i].C)
	}
	return nil
}

func (s *Store) put(j ffs.Job) error {
	buf, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("marshaling for datastore: %s", err)
	}
	if err := s.ds.Put(makeKey(j.ID), buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}
	s.notifyWatchers(j)
	if j.Status == ffs.InProgress {
		s.inProgress[j.Cid] = struct{}{}
	} else if j.Status == ffs.Failed || j.Status == ffs.Success {
		delete(s.inProgress, j.Cid)
	}
	return nil
}

func (s *Store) get(jid ffs.JobID) (ffs.Job, error) {
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return ffs.Job{}, scheduler.ErrNotFound
	}
	if err != nil {
		return ffs.Job{}, fmt.Errorf("getting job from datastore: %s", err)
	}
	var job ffs.Job
	if err := json.Unmarshal(buf, &job); err != nil {
		return job, fmt.Errorf("unmarshaling job from datastore: %s", err)
	}
	return job, nil
}

func (s *Store) notifyWatchers(j ffs.Job) {
	for _, w := range s.watchers {
		if w.iid != j.APIID {
			continue
		}
		select {
		case w.C <- j:
			log.Infof("notifying watcher")
		default:
			log.Warnf("slow watcher skipped job %s", j.ID)
		}
	}
}

func makeKey(jid ffs.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
