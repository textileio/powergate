package jstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("ffs-sched-jstore")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("job not found")

	dsBaseJob          = datastore.NewKey("job")
	dsBaseStartedDeals = datastore.NewKey("starteddeals")
)

// Store is a Datastore implementation of JobStore, which saves
// state of scheduler Jobs.
type Store struct {
	lock     sync.Mutex
	ds       datastore.Datastore
	watchers []watcher

	executingCids map[cid.Cid]ffs.JobID
}

type watcher struct {
	iid ffs.APIID
	C   chan ffs.Job
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) (*Store, error) {
	s := &Store{
		ds:            ds,
		executingCids: make(map[cid.Cid]ffs.JobID),
	}
	if err := s.loadExecutingJobs(); err != nil {
		return nil, fmt.Errorf("reloading priori executing jobs: %s", err)
	}
	return s, nil
}

// Finalize sets a Job status to a final state, i.e. Success or Failed,
// with a list of Deal errors occurred during job execution.
func (s *Store) Finalize(jid ffs.JobID, st ffs.JobStatus, jobError error, dealErrors []ffs.DealError) error {
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
	j.DealErrors = dealErrors
	if err := s.put(j); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// Dequeue dequeues a Job which doesn't have have another Executing Job
// for the same Cid. Saying it differently, it's safe to execute. The returned
// job Status is automatically changed to Executing. If no jobs are available to dequeue
// it returns a nil *ffs.Job and no-error.
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
		_, ok := s.executingCids[j.Cid]
		if j.Status == ffs.Queued && !ok {
			j.Status = ffs.Executing
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

// GetExecutingJobs returns the JobIDs of all Jobs in Executing status.
func (s *Store) GetExecutingJobs() []ffs.JobID {
	s.lock.Lock()
	defer s.lock.Unlock()
	res := make([]ffs.JobID, len(s.executingCids))
	var i int
	for _, jid := range s.executingCids {
		res[i] = jid
		i++
	}
	return res
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

// StartedDeals describe deals that are currently waiting to have a
// final status.
type StartedDeals struct {
	Cid          cid.Cid
	ProposalCids []cid.Cid
}

// AddStartedDeals is a temporal storage solution of deals that are started
// are being watched. It serves as a recovery point to reattach to fired
// deals when the scheduler was abruptly interrupted.
func (s *Store) AddStartedDeals(c cid.Cid, proposals []cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	sd := StartedDeals{Cid: c, ProposalCids: proposals}
	buf, err := json.Marshal(sd)
	if err != nil {
		return fmt.Errorf("marshaling started deals: %s", err)
	}
	if err := s.ds.Put(makeStartedDealsKey(c), buf); err != nil {
		return fmt.Errorf("saving started deals to datastore: %s", err)
	}
	return nil
}

// RemoveStartedDeals removes all started deals from Cid.
func (s *Store) RemoveStartedDeals(c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ds.Delete(makeStartedDealsKey(c)); err != nil {
		return fmt.Errorf("deleting started deals from datastore: %s", err)
	}
	return nil
}

// GetStartedDeals gets all stored started deals from Cid. If no started
// deals are present, an empty slice is returned.
func (s *Store) GetStartedDeals(c cid.Cid) ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var sd StartedDeals
	b, err := s.ds.Get(makeStartedDealsKey(c))
	if err == datastore.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting started deals from datastore: %s", err)
	}
	if err := json.Unmarshal(b, &sd); err != nil {
		return nil, fmt.Errorf("unmarshaling started deals from datastore: %s", err)
	}
	return sd.ProposalCids, nil
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
	if j.Status == ffs.Executing {
		s.executingCids[j.Cid] = j.ID
	} else if j.Status == ffs.Failed || j.Status == ffs.Success {
		delete(s.executingCids, j.Cid)
	}
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

func (s *Store) loadExecutingJobs() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{Prefix: dsBaseJob.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying executing jobs in datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()
	for r := range res.Next() {
		var j ffs.Job
		if err := json.Unmarshal(r.Value, &j); err != nil {
			return fmt.Errorf("unmarshalling job: %s", err)
		}
		if j.Status == ffs.Executing {
			s.executingCids[j.Cid] = j.ID
		}
	}
	return nil
}

func makeStartedDealsKey(c cid.Cid) datastore.Key {
	return dsBaseStartedDeals.ChildString(util.CidToString(c))
}

func makeKey(jid ffs.JobID) datastore.Key {
	return dsBaseJob.ChildString(jid.String())
}
