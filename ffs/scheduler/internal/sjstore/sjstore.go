package sjstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("ffs-sched-sjstore")

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

	queued        []ffs.StorageJob
	executingCids map[cid.Cid]ffs.JobID

	jobStatusCache map[ffs.APIID]map[cid.Cid]map[cid.Cid]deals.StorageDealInfo

	queuedJobs         map[ffs.APIID]map[cid.Cid][]*ffs.StorageJob
	executingJobs      map[ffs.APIID]map[cid.Cid]*ffs.StorageJob
	lastFinalJobs      map[ffs.APIID]map[cid.Cid]*ffs.StorageJob
	lastSuccessfulJobs map[ffs.APIID]map[cid.Cid]*ffs.StorageJob
}

// Stats return metrics about current job queues.
type Stats struct {
	TotalQueued    int
	TotalExecuting int
}

type watcher struct {
	iid ffs.APIID
	C   chan ffs.StorageJob
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) (*Store, error) {
	s := &Store{
		ds:                 ds,
		executingCids:      make(map[cid.Cid]ffs.JobID),
		jobStatusCache:     make(map[ffs.APIID]map[cid.Cid]map[cid.Cid]deals.StorageDealInfo),
		queuedJobs:         make(map[ffs.APIID]map[cid.Cid][]*ffs.StorageJob),
		executingJobs:      make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
		lastFinalJobs:      make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
		lastSuccessfulJobs: make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
	}
	if err := s.loadCaches(); err != nil {
		return nil, fmt.Errorf("reloading caches: %s", err)
	}
	return s, nil
}

// MonitorJob returns a channel that can be passed into the deal monitoring process.
func (s *Store) MonitorJob(j ffs.StorageJob) chan deals.StorageDealInfo {
	dealUpdates := make(chan deals.StorageDealInfo, 1000)
	go func() {
		for update := range dealUpdates {
			s.lock.Lock()
			_, ok := s.jobStatusCache[j.APIID]
			if !ok {
				s.jobStatusCache[j.APIID] = map[cid.Cid]map[cid.Cid]deals.StorageDealInfo{}
			}
			_, ok = s.jobStatusCache[j.APIID][j.Cid]
			if !ok {
				s.jobStatusCache[j.APIID][j.Cid] = map[cid.Cid]deals.StorageDealInfo{}
			}
			s.jobStatusCache[j.APIID][j.Cid][update.ProposalCid] = update
			job, err := s.get(j.ID)
			if err != nil {
				log.Errorf("getting job: %v", err)
				s.lock.Unlock()
				continue
			}
			values := make([]deals.StorageDealInfo, 0, len(s.jobStatusCache[j.APIID][j.Cid]))
			for _, v := range s.jobStatusCache[j.APIID][j.Cid] {
				values = append(values, v)
			}
			sort.Slice(values, func(i, j int) bool {
				return values[i].ProposalCid.String() < values[j].ProposalCid.String()
			})
			job.DealInfo = values
			if err := s.put(job); err != nil {
				log.Errorf("saving job with deal info updates: %v", err)
			}
			s.lock.Unlock()
		}
		s.lock.Lock()
		delete(s.jobStatusCache[j.APIID], j.Cid)
		s.lock.Unlock()
	}()
	return dealUpdates
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
func (s *Store) Dequeue() (*ffs.StorageJob, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, job := range s.queued {
		execJobID, ok := s.executingCids[job.Cid]
		if job.Status == ffs.Queued && !ok {
			job.Status = ffs.Executing
			if err := s.put(job); err != nil {
				return nil, err
			}
			return &job, nil
		}
		log.Infof("queued %s is delayed since job %s is running", job.ID, execJobID)
	}
	return nil, nil
}

// Enqueue queues a new Job. If other Job for the same Cid is in Queued status,
// it will be automatically marked as Canceled.
func (s *Store) Enqueue(j ffs.StorageJob) error {
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

// GetExecutingJob returns a JobID that is currently executing for
// data with cid c. If there's not such job, it returns nil.
func (s *Store) GetExecutingJob(c cid.Cid) *ffs.JobID {
	j, ok := s.executingCids[c]
	if !ok {
		return nil
	}
	return &j
}

// GetStats return the current Stats for storage jobs.
func (s *Store) GetStats() Stats {
	s.lock.Lock()
	defer s.lock.Unlock()
	return Stats{
		TotalQueued:    len(s.queued),
		TotalExecuting: len(s.executingCids),
	}
}

// GetExecutingJobIDs returns the JobIDs of all Jobs in Executing status.
func (s *Store) GetExecutingJobIDs() []ffs.JobID {
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
			log.Errorf("closing cancelQueued query result: %s", err)
		}
	}()
	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iter next: %s", r.Error)
		}
		var j ffs.StorageJob
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
func (s *Store) Get(jid ffs.JobID) (ffs.StorageJob, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.get(jid)
}

// Watch subscribes to Job changes from a specified Api instance.
func (s *Store) Watch(ctx context.Context, c chan<- ffs.StorageJob, iid ffs.APIID) error {
	s.lock.Lock()
	ic := make(chan ffs.StorageJob, 1)
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

// QueuedJobs returns queued jobs for the specified instance id and cids.
// If the instance id is ffs.EmptyInstanceID, data for all instances is returned.
// If no cids are provided, data for all data cids is returned.
func (s *Store) QueuedJobs(iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob {
	s.lock.Lock()
	defer s.lock.Unlock()

	var iids []ffs.APIID
	if iid == ffs.EmptyInstanceID {
		for iid := range s.queuedJobs {
			iids = append(iids, iid)
		}
	} else {
		iids = append(iids, iid)
		ensureJobsSliceMap(s.queuedJobs, iid)
	}

	var res []ffs.StorageJob
	for _, iid := range iids {
		cidsList := cids
		if len(cidsList) == 0 {
			for cid := range s.queuedJobs[iid] {
				cidsList = append(cidsList, cid)
			}
		}
		for _, cid := range cidsList {
			jobs := s.queuedJobs[iid][cid]
			for _, job := range jobs {
				res = append(res, *job)
			}
		}
	}

	sort.Slice(res, func(a, b int) bool {
		return res[a].CreatedAt < res[b].CreatedAt
	})

	return res
}

// ExecutingJobs returns executing jobs for the specified instance id and cids.
// If the instance id is ffs.EmptyInstanceID, data for all instances is returned.
// If no cids are provided, data for all data cids is returned.
func (s *Store) ExecutingJobs(iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob {
	s.lock.Lock()
	defer s.lock.Unlock()

	return mappedJobs(s.executingJobs, iid, cids...)
}

// LatestFinalJobs returns the most recent finished jobs for the specified instance id and cids.
// If the instance id is ffs.EmptyInstanceID, data for all instances is returned.
// If no cids are provided, data for all data cids is returned.
func (s *Store) LatestFinalJobs(iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob {
	s.lock.Lock()
	defer s.lock.Unlock()

	return mappedJobs(s.lastFinalJobs, iid, cids...)
}

// LatestSuccessfulJobs returns the most recent successful jobs for the specified instance id and cids.
// If the instance id is ffs.EmptyInstanceID, data for all instances is returned.
// If no cids are provided, data for all data cids is returned.
func (s *Store) LatestSuccessfulJobs(iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob {
	s.lock.Lock()
	defer s.lock.Unlock()

	return mappedJobs(s.lastSuccessfulJobs, iid, cids...)
}

func mappedJobs(m map[ffs.APIID]map[cid.Cid]*ffs.StorageJob, iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob {
	var iids []ffs.APIID
	if iid == ffs.EmptyInstanceID {
		for iid := range m {
			iids = append(iids, iid)
		}
	} else {
		iids = append(iids, iid)
		ensureJobsMap(m, iid)
	}

	var res []ffs.StorageJob
	for _, iid := range iids {
		cidsList := cids
		if len(cidsList) == 0 {
			for cid := range m[iid] {
				cidsList = append(cidsList, cid)
			}
		}

		for _, cid := range cidsList {
			job := m[iid][cid]
			if job != nil {
				res = append(res, *job)
			}
		}
	}
	return res
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

func (s *Store) put(j ffs.StorageJob) error {
	buf, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("marshaling for datastore: %s", err)
	}
	if err := s.ds.Put(makeKey(j.ID), buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}

	// Update executing cids cache.
	ensureJobsMap(s.executingJobs, j.APIID)
	if j.Status == ffs.Executing {
		s.executingCids[j.Cid] = j.ID
		s.executingJobs[j.APIID][j.Cid] = &j
	} else {
		execJobID, ok := s.executingCids[j.Cid]
		// If the executing job is not longer executing,
		// take out from executing cache.
		if ok && execJobID == j.ID {
			delete(s.executingCids, j.Cid)
		}

		execJob, ok := s.executingJobs[j.APIID][j.Cid]
		if ok && execJob.ID == j.ID {
			delete(s.executingJobs[j.APIID], j.Cid)
		}
	}

	// Update queued cids cache.
	ensureJobsSliceMap(s.queuedJobs, j.APIID)
	if j.Status == ffs.Queued {
		// Every new Queued job goes at the tail since has
		// the biggest CreatedAt value.
		s.queued = append(s.queued, j)
		s.queuedJobs[j.APIID][j.Cid] = append(s.queuedJobs[j.APIID][j.Cid], &j)
	} else { // In any other case, ensure taking it out from queued caches.
		delIndex := -1
		for i, job := range s.queued {
			if j.ID == job.ID {
				delIndex = i
				break
			}
		}
		if delIndex != -1 {
			s.queued = append(s.queued[:delIndex], s.queued[delIndex+1:]...)
		}

		delIndex = -1
		for i, job := range s.queuedJobs[j.APIID][j.Cid] {
			if j.ID == job.ID {
				delIndex = i
				break
			}
		}
		if delIndex != -1 {
			s.queuedJobs[j.APIID][j.Cid] = append(s.queuedJobs[j.APIID][j.Cid][:delIndex], s.queuedJobs[j.APIID][j.Cid][delIndex+1:]...)
		}
	}

	// Update the cache of latest final jobs
	if isFinal(j) {
		ensureJobsMap(s.lastFinalJobs, j.APIID)
		s.lastFinalJobs[j.APIID][j.Cid] = &j
	}

	// Update the cache of latest successful jobs
	if isSuccessful(j) {
		ensureJobsMap(s.lastSuccessfulJobs, j.APIID)
		s.lastSuccessfulJobs[j.APIID][j.Cid] = &j
	}

	s.notifyWatchers(j)
	return nil
}

func (s *Store) get(jid ffs.JobID) (ffs.StorageJob, error) {
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return ffs.StorageJob{}, ErrNotFound
	}
	if err != nil {
		return ffs.StorageJob{}, fmt.Errorf("getting job from datastore: %s", err)
	}
	var job ffs.StorageJob
	if err := json.Unmarshal(buf, &job); err != nil {
		return job, fmt.Errorf("unmarshaling job from datastore: %s", err)
	}
	return job, nil
}

func (s *Store) notifyWatchers(j ffs.StorageJob) {
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

func (s *Store) loadCaches() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := query.Query{Prefix: dsBaseJob.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing queued jobs result: %s", err)
		}
	}()
	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iter next: %s", r.Error)
		}
		var j ffs.StorageJob
		if err := json.Unmarshal(r.Value, &j); err != nil {
			return fmt.Errorf("unmarshalling job: %s", err)
		}

		if j.Status == ffs.Queued {
			s.queued = append(s.queued, j)
			ensureJobsSliceMap(s.queuedJobs, j.APIID)
			s.queuedJobs[j.APIID][j.Cid] = append(s.queuedJobs[j.APIID][j.Cid], &j)
		} else if j.Status == ffs.Executing {
			s.executingCids[j.Cid] = j.ID
			ensureJobsMap(s.executingJobs, j.APIID)
			s.executingJobs[j.APIID][j.Cid] = &j
		}

		ensureJobsMap(s.lastFinalJobs, j.APIID)
		newest, ok := s.lastFinalJobs[j.APIID][j.Cid]
		if isFinal(j) && (!ok || j.CreatedAt > newest.CreatedAt) {
			s.lastFinalJobs[j.APIID][j.Cid] = &j
		}

		ensureJobsMap(s.lastSuccessfulJobs, j.APIID)
		newest, ok = s.lastSuccessfulJobs[j.APIID][j.Cid]
		if isSuccessful(j) && (!ok || j.CreatedAt > newest.CreatedAt) {
			s.lastSuccessfulJobs[j.APIID][j.Cid] = &j
		}
	}
	sort.Slice(s.queued, func(a, b int) bool {
		return s.queued[a].CreatedAt < s.queued[b].CreatedAt
	})
	for _, cidMap := range s.queuedJobs {
		for _, queuedJobs := range cidMap {
			sort.Slice(queuedJobs, func(a, b int) bool {
				return queuedJobs[a].CreatedAt < queuedJobs[b].CreatedAt
			})
		}
	}
	return nil
}

func ensureJobsMap(m map[ffs.APIID]map[cid.Cid]*ffs.StorageJob, apiID ffs.APIID) {
	_, ok := m[apiID]
	if !ok {
		m[apiID] = map[cid.Cid]*ffs.StorageJob{}
	}
}

func ensureJobsSliceMap(m map[ffs.APIID]map[cid.Cid][]*ffs.StorageJob, apiID ffs.APIID) {
	_, ok := m[apiID]
	if !ok {
		m[apiID] = map[cid.Cid][]*ffs.StorageJob{}
	}
}

func isFinal(j ffs.StorageJob) bool {
	return j.Status == ffs.Success || j.Status == ffs.Failed || j.Status == ffs.Canceled
}

func isSuccessful(j ffs.StorageJob) bool {
	return j.Status == ffs.Success
}

func makeStartedDealsKey(c cid.Cid) datastore.Key {
	return dsBaseStartedDeals.ChildString(util.CidToString(c))
}

func makeKey(jid ffs.JobID) datastore.Key {
	return dsBaseJob.ChildString(jid.String())
}
