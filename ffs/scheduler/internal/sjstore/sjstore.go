package sjstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

/**
There are many namespaces that are maintained in the contained datastore.
Use these descriptions to understand what information is stored where.
Might be useful in a future migration or datastore maintenance.

/job/<job-id>: Stores StorageJob data by job-id
/apiid/<api-id>/<cid>/<timestamp>: Index on api-id primarily, cid secondarily, with timestamp, values of job-id
/cid/<cid>/<api-id>/<timestamp>: Index on cid primarily, api-id secondarily, with timestamp, values of job-id
/starteddeals_v2/<instance-id>/<cid>: Stores StartedDeals data by instance-id and cid
*/

var (
	log = logging.Logger("ffs-sched-sjstore")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("job not found")

	dsBaseJob          = datastore.NewKey("job")
	dsBaseAPIID        = datastore.NewKey("apiid")
	dsBaseCid          = datastore.NewKey("cid")
	dsBaseStartedDeals = datastore.NewKey("starteddeals_v2")
)

// Store is a Datastore implementation of JobStore, which saves
// state of scheduler Jobs.
type Store struct {
	lock     sync.Mutex
	ds       datastore.TxnDatastore
	watchers []watcher

	queued []ffs.StorageJob

	jobStatusCache map[ffs.APIID]map[cid.Cid]map[cid.Cid]deals.StorageDealInfo

	queuedJobs         map[ffs.APIID]map[cid.Cid][]*ffs.StorageJob
	executingJobs      map[ffs.APIID]map[cid.Cid]*ffs.StorageJob
	lastFinalJobs      map[ffs.APIID]map[cid.Cid]*ffs.StorageJob
	lastSuccessfulJobs map[ffs.APIID]map[cid.Cid]*ffs.StorageJob

	queuedIDs    map[ffs.JobID]struct{}
	executingIDs map[ffs.JobID]struct{}
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
func New(ds datastore.TxnDatastore) (*Store, error) {
	s := &Store{
		ds:                 ds,
		jobStatusCache:     make(map[ffs.APIID]map[cid.Cid]map[cid.Cid]deals.StorageDealInfo),
		queuedJobs:         make(map[ffs.APIID]map[cid.Cid][]*ffs.StorageJob),
		executingJobs:      make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
		lastFinalJobs:      make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
		lastSuccessfulJobs: make(map[ffs.APIID]map[cid.Cid]*ffs.StorageJob),
		queuedIDs:          make(map[ffs.JobID]struct{}),
		executingIDs:       make(map[ffs.JobID]struct{}),
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
			if err := s.put(job, false); err != nil {
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
	if err := s.put(j, false); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// Dequeue dequeues a Job which doesn't have have another Executing Job
// for the same Cid. Saying it differently, it's safe to execute. The returned
// job Status is automatically changed to Executing. If an instance id is provided,
// only a job for that instance id will be dequeued. If no jobs are available to dequeue
// it returns a nil *ffs.Job and no-error.
func (s *Store) Dequeue(iid ffs.APIID) (*ffs.StorageJob, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, job := range s.queued {
		execJob, ok := s.executingJobs[job.APIID][job.Cid]
		isAPIIDMatch := true
		if iid != ffs.EmptyInstanceID {
			isAPIIDMatch = iid == job.APIID
		}
		if job.Status == ffs.Queued && !ok && isAPIIDMatch {
			job.Status = ffs.Executing
			if err := s.put(job, false); err != nil {
				return nil, err
			}
			return &job, nil
		}
		if ok {
			// ToDo: Maybe remove this since there might be lots of reasons we skip over a job.
			// For example, if the specified iid doesn't match, but that is not worth logging.
			log.Infof("queued %s is delayed since job %s is running", job.ID, execJob.ID)
		}
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
	if err := s.put(j, true); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}

	return nil
}

// GetExecutingJob returns a JobID that is currently executing for
// data with cid c in iid. If there's not such job, it returns nil.
func (s *Store) GetExecutingJob(iid ffs.APIID, c cid.Cid) *ffs.JobID {
	j, ok := s.executingJobs[iid][c]
	if !ok {
		return nil
	}
	return &j.ID
}

// GetStats return the current Stats for storage jobs.
func (s *Store) GetStats() Stats {
	s.lock.Lock()
	defer s.lock.Unlock()
	var count int
	for _, iidJobs := range s.executingJobs {
		count += len(iidJobs)
	}
	return Stats{
		TotalQueued:    len(s.queued),
		TotalExecuting: count,
	}
}

// GetExecutingJobIDs returns the JobIDs of all Jobs in Executing status.
func (s *Store) GetExecutingJobIDs() []ffs.JobID {
	s.lock.Lock()
	defer s.lock.Unlock()
	var res []ffs.JobID
	for _, iid := range s.executingJobs {
		for _, j := range iid {
			res = append(res, j.ID)
		}
	}
	return res
}

// CancelQueued cancels a job if it's in Queued status.
// If the Job isn't Queued, the call is a noop. If the Job
// doesn't exist it returns ErrNotFound.
func (s *Store) CancelQueued(jid ffs.JobID) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	j, err := s.get(jid)
	if err != nil {
		return false, err
	}
	if j.Status != ffs.Queued {
		return false, nil
	}
	j.Status = ffs.Canceled
	if err := s.put(j, false); err != nil {
		return false, fmt.Errorf("canceling queued job: %s", err)
	}

	return true, nil
}

func (s *Store) cancelQueued(c cid.Cid) error {
	q := query.Query{Prefix: dsBaseJob.String()}
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
			if err := s.put(j, false); err != nil {
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
func (s *Store) AddStartedDeals(iid ffs.APIID, c cid.Cid, proposals []cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	sd := StartedDeals{Cid: c, ProposalCids: proposals}
	buf, err := json.Marshal(sd)
	if err != nil {
		return fmt.Errorf("marshaling started deals: %s", err)
	}
	if err := s.ds.Put(makeStartedDealsKey(iid, c), buf); err != nil {
		return fmt.Errorf("saving started deals to datastore: %s", err)
	}
	return nil
}

// RemoveStartedDeals removes all started deals from Cid.
func (s *Store) RemoveStartedDeals(iid ffs.APIID, c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ds.Delete(makeStartedDealsKey(iid, c)); err != nil {
		return fmt.Errorf("deleting started deals from datastore: %s", err)
	}
	return nil
}

// GetStartedDeals gets all stored started deals from Cid for an APIID.
// If no started deals are present, an empty slice is returned.
func (s *Store) GetStartedDeals(iid ffs.APIID, c cid.Cid) ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	b, err := s.ds.Get(makeStartedDealsKey(iid, c))
	if err == datastore.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting started deals from datastore: %s", err)
	}
	var sd StartedDeals
	if err := json.Unmarshal(b, &sd); err != nil {
		return nil, fmt.Errorf("unmarshaling started deals from datastore: %s", err)
	}
	return sd.ProposalCids, nil
}

// Select specifies which StorageJobs to list.
type Select int

const (
	// All lists all StorageJobs and is the default.
	All Select = iota
	// Queued lists queued StorageJobs.
	Queued
	// Executing lists executing StorageJobs.
	Executing
	// Final lists final StorageJobs.
	Final
)

// ListConfig controls the behavior for listing StorageJobs.
type ListConfig struct {
	// APIIDFilter filters StorageJobs list to the specified APIID. Defaults to no filter.
	APIIDFilter ffs.APIID
	// CidFilter filters StorageJobs list to the specified cid. Defaults to no filter.
	CidFilter cid.Cid
	// Limit limits the number of StorageJobs returned. Defaults to no limit.
	Limit uint64
	// Ascending returns the StorageJobs ascending by time. Defaults to false, descending.
	Ascending bool
	// Select specifies to return StorageJobs in the specified state.
	Select Select
	// NextPageToken sets the slug from which to start building the next page of results.
	NextPageToken string
}

// List lists StorageJobs according to the provided ListConfig.
func (s *Store) List(config ListConfig) ([]ffs.StorageJob, bool, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	byTime := func(ascending bool) func(a, b query.Entry) int {
		extractTime := func(key string) (int64, error) {
			parts := strings.Split(key, "/")
			if len(parts) != 5 {
				return 0, fmt.Errorf("expected 5 key parts but got %v", len(parts))
			}
			return strconv.ParseInt(parts[4], 10, 64)
		}
		return func(a, b query.Entry) int {
			l := a
			r := b
			if ascending {
				l = b
				r = a
			}
			lTime, err := extractTime(l.Key)
			if err != nil {
				log.Errorf("extracting time from key a: %v", err)
				return 0
			}
			rTime, err := extractTime(r.Key)
			if err != nil {
				log.Errorf("extracting time from key b: %v", err)
				return 0
			}
			if lTime > rTime {
				return -1
			} else if rTime > lTime {
				return 1
			} else {
				return 0
			}
		}
	}

	// by default, query all.
	prefix := dsBaseCid.String()
	if config.APIIDFilter != ffs.EmptyInstanceID && config.CidFilter.Defined() {
		// use apiid/cid
		prefix = prefixAPIIDAndCid(config.APIIDFilter, config.CidFilter)
	} else if config.APIIDFilter != ffs.EmptyInstanceID && !config.CidFilter.Defined() {
		// use apiid
		prefix = prefixAPIID(config.APIIDFilter)
	} else if config.APIIDFilter == ffs.EmptyInstanceID && config.CidFilter.Defined() {
		// use cid
		prefix = prefixCid(config.CidFilter)
	}

	q := query.Query{
		Prefix: prefix,
		Orders: []query.Order{query.OrderByFunction(byTime(config.Ascending))},
	}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, false, "", fmt.Errorf("querying datastore: %v", err)
	}
	var jobs []ffs.StorageJob
	foundNextPageToken := false
	if config.NextPageToken == "" {
		foundNextPageToken = true
	}
	done := false
	more := false
	nextPageToken := ""
	for r := range res.Next() {
		// return an error if there was an error iterating next.
		if r.Error != nil {
			return nil, false, "", fmt.Errorf("iter next: %s", r.Error)
		}

		// if in the last loop we decided we're done, we use this iteration
		// just to note that there is more data available then break.
		if done {
			more = true
			break
		}

		jobIDString := string(r.Value)
		jobID := ffs.JobID(jobIDString)

		// if we haven't found the record we need to seek to, continue to the next.
		if !foundNextPageToken {
			// additionally, if this is the record we are seeking to, note that we've found it, then continue.
			if config.NextPageToken == jobIDString {
				foundNextPageToken = true
			}
			continue
		}

		// Filter out based on queued/executing etc
		switch config.Select {
		case All:
		case Queued:
			if _, queued := s.queuedIDs[jobID]; !queued {
				continue
			}
		case Executing:
			if _, executing := s.executingIDs[jobID]; !executing {
				continue
			}
		case Final:
			_, queued := s.queuedIDs[jobID]
			_, executing := s.executingIDs[jobID]
			if queued || executing {
				continue
			}
		}

		job, err := s.get(jobID)
		if err != nil {
			return nil, false, "", fmt.Errorf("getting job: %v", err)
		}
		jobs = append(jobs, job)
		nextPageToken = jobIDString
		if len(jobs) == int(config.Limit) {
			done = true
		}
	}

	if !more {
		nextPageToken = ""
	}

	return jobs, more, nextPageToken, nil
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

func (s *Store) put(j ffs.StorageJob, updateIndex bool) error {
	buf, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("marshaling for datastore: %s", err)
	}

	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("starting transaction: %s", err)
	}
	defer txn.Discard()

	if err := s.ds.Put(makeKey(j.ID), buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}

	if updateIndex {
		if err := s.ds.Put(makeAPIIDKey(j), []byte(j.ID)); err != nil {
			return fmt.Errorf("saving to api id index: %s", err)
		}
		if err := s.ds.Put(makeCidKey(j), []byte(j.ID)); err != nil {
			return fmt.Errorf("saving to cid index: %s", err)
		}
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("committing txn: %v", err)
	}

	// Update executing cids cache.
	ensureJobsMap(s.executingJobs, j.APIID)
	if j.Status == ffs.Executing {
		s.executingJobs[j.APIID][j.Cid] = &j
		s.executingIDs[j.ID] = struct{}{}
	} else {
		execJob, ok := s.executingJobs[j.APIID][j.Cid]
		if ok && execJob.ID == j.ID {
			delete(s.executingJobs[j.APIID], j.Cid)
		}
		delete(s.executingIDs, j.ID)
	}

	// Update queued cids cache.
	ensureJobsSliceMap(s.queuedJobs, j.APIID)
	if j.Status == ffs.Queued {
		// Every new Queued job goes at the tail since has
		// the biggest CreatedAt value.
		s.queued = append(s.queued, j)
		s.queuedJobs[j.APIID][j.Cid] = append(s.queuedJobs[j.APIID][j.Cid], &j)
		s.queuedIDs[j.ID] = struct{}{}
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
		delete(s.queuedIDs, j.ID)
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
			s.queuedIDs[j.ID] = struct{}{}
		} else if j.Status == ffs.Executing {
			ensureJobsMap(s.executingJobs, j.APIID)
			s.executingJobs[j.APIID][j.Cid] = &j
			s.executingIDs[j.ID] = struct{}{}
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

func makeStartedDealsKey(iid ffs.APIID, c cid.Cid) datastore.Key {
	return dsBaseStartedDeals.ChildString(iid.String()).ChildString(util.CidToString(c))
}

func makeKey(jid ffs.JobID) datastore.Key {
	return dsBaseJob.ChildString(jid.String())
}

func makeAPIIDKey(j ffs.StorageJob) datastore.Key {
	return datastore.NewKey(prefixAPIIDAndCid(j.APIID, j.Cid)).ChildString(fmt.Sprintf("%d", j.CreatedAt))
}

func makeCidKey(j ffs.StorageJob) datastore.Key {
	return datastore.NewKey(prefixCidAndAPIID(j.Cid, j.APIID)).ChildString(fmt.Sprintf("%d", j.CreatedAt))
}

func prefixCid(cid cid.Cid) string {
	return dsBaseCid.ChildString(cid.String()).String()
}

func prefixAPIID(APIID ffs.APIID) string {
	return dsBaseAPIID.ChildString(APIID.String()).String()
}

func prefixCidAndAPIID(cid cid.Cid, APIID ffs.APIID) string {
	return dsBaseCid.ChildString(cid.String()).ChildString(APIID.String()).String()
}

func prefixAPIIDAndCid(APIID ffs.APIID, cid cid.Cid) string {
	return dsBaseAPIID.ChildString(APIID.String()).ChildString(cid.String()).String()
}
