package jstore

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sched-jstore")
)

// JobStore is an scheduler.JobStore implementation that saves
// state of scheduler Jobs in a Datastore.
type Store struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher
}

var _ scheduler.JobStore = (*Store)(nil)

type watcher struct {
	iid ffs.ApiID
	ch  chan ffs.Job
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// GetByStatus returns all Jobs with the specified JobStatus.
func (s *Store) GetByStatus(status ffs.JobStatus) ([]ffs.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := query.Query{Prefix: ""}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying datastore: %s", err)
	}
	defer res.Close()

	var ret []ffs.Job
	for r := range res.Next() {
		var job ffs.Job
		if err := json.Unmarshal(r.Value, &job); err != nil {
			return nil, fmt.Errorf("unmarshalling job in query: %s", err)
		}
		if job.Status == status {
			ret = append(ret, job)
		}
	}
	return ret, nil
}

// Put saves Job's data in the Datastore.
func (s *Store) Put(j ffs.Job) error {
	s.lock.Lock()
	defer s.lock.Unlock()
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

// Get returns the current state of Job. If doesn't exist, returns
// ErrNotFound.
func (s *Store) Get(jid ffs.JobID) (ffs.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

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

// Watch subscribes to Job changes from a specified Api instance.
func (s *Store) Watch(iid ffs.ApiID) <-chan ffs.Job {
	s.lock.Lock()
	defer s.lock.Unlock()

	ch := make(chan ffs.Job, 1)
	s.watchers = append(s.watchers, watcher{iid: iid, ch: ch})
	return ch
}

// Unwatch unregisters a channel returned from Watch().
func (s *Store) Unwatch(ch <-chan ffs.Job) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i := range s.watchers {
		if s.watchers[i].ch == ch {
			close(s.watchers[i].ch)
			s.watchers[i] = s.watchers[len(s.watchers)-1]
			s.watchers = s.watchers[:len(s.watchers)-1]
		}
	}
}

// Close closes the Store, unregistering any subscribed watchers.
func (s *Store) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i := range s.watchers {
		close(s.watchers[i].ch)
	}
	return nil
}

func (s *Store) notifyWatchers(j ffs.Job) {
	for _, w := range s.watchers {
		if w.iid != j.InstanceID {
			continue
		}
		select {
		case w.ch <- j:
			log.Infof("notifying watcher")
		default:
			log.Warnf("slow watcher skipped job %s", j.ID)
		}
	}
}

func makeKey(jid ffs.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
