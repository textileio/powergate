package rjstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("ffs-sched-rjstore")

	// ErrNotFound indicates the job doesn't exists.
	ErrNotFound = errors.New("retrieval job not found")
	dsBaseJob   = datastore.NewKey("job")
)

// Store is a persistent store for Retrieval Jobs.
type Store struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher
}

// watcher represents an API instance who is watching for
// Job updates.
type watcher struct {
	iid ffs.APIID
	C   chan ffs.Job
}

// New returns a new retrieval Job store.
func New(ds datastore.Datastore) (*Store, error) {
	s := &Store{ds: ds}
	return s, nil
}

// Finalize sets a Job status to a final state.
func (s *Store) Finalize(jid ffs.JobID, st ffs.JobStatus, jobError error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	j, err := s.get(jid)
	if err != nil {
		return err
	}
	switch st {
	case ffs.Success, ffs.Failed, ffs.Canceled:
		// Success: Job executed within expected behavior.
		// Failed: Job executed with expected failure scenario.
		// Canceled: Job was canceled by the client.
	default:
		return fmt.Errorf("can't finalize job with status %s", ffs.JobStatusStr[st])
	}
	j.Status = st
	if jobError != nil {
		j.ErrCause = jobError.Error()
	}
	if err := s.put(j); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// Dequeue dequeues a ready to be executed Job.
func (s *Store) Dequeue() (*ffs.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{Prefix: dsBaseJob.String()}
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
		if j.Status == ffs.Queued {
			j.Status = ffs.Executing
			if err := s.put(j); err != nil {
				return nil, err
			}
			return &j, nil
		}
	}
	return nil, nil
}

// Enqueue queues a new Job.
func (s *Store) Enqueue(j ffs.Job) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	j.Status = ffs.Queued
	if err := s.put(j); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
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
	s.watchers = nil
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
	return nil
}

func (s *Store) get(jid ffs.JobID) (ffs.Job, error) {
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return ffs.Job{}, ErrNotFound
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
	return dsBaseJob.ChildString(jid.String())
}
