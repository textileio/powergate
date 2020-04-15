package jstore

import (
	"context"
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

// Store is a Datastore implementation of JobStore, which saves
// state of scheduler Jobs.
type Store struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher
}

var _ scheduler.JobStore = (*Store)(nil)

type watcher struct {
	iid ffs.APIID
	C   chan ffs.Job
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
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing getbystatus query result: %s", err)
		}
	}()

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

func (s *Store) notifyWatchers(j ffs.Job) {
	for _, w := range s.watchers {
		if w.iid != j.InstanceID {
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
