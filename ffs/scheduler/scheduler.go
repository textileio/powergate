package scheduler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/astore"
	"github.com/textileio/powergate/ffs/scheduler/internal/cistore"
	"github.com/textileio/powergate/ffs/scheduler/internal/jstore"
	"github.com/textileio/powergate/ffs/scheduler/internal/trackstore"
	txndstr "github.com/textileio/powergate/txndstransform"
)

const (
	maxParallelExecutions = 50
)

var (
	log = logging.Logger("ffs-scheduler")

	// ErrNotFound is returned when an item isn't found on a Store.
	ErrNotFound = errors.New("item not found")

	// RenewalEvalFrequency is the frequency in which renewable StorageConfigs
	// will be evaluated.
	RenewalEvalFrequency = time.Hour * 24

	// RepairEvalFrequency is the frequency in which repairable StorageConfigs
	// will be evaluated.
	RepairEvalFrequency = time.Hour * 24
)

// Scheduler receives actions to store a Cid in Hot and Cold layers. These actions are
// created as Jobs which have a lifecycle that can be watched by external actors.
// This Jobs are executed by delegating the work to the Hot and Cold layers configured for
// the scheduler.
type Scheduler struct {
	cs  ffs.ColdStorage
	hs  ffs.HotStorage
	js  *jstore.Store
	as  *astore.Store
	ts  *trackstore.Store
	cis *cistore.Store
	l   ffs.CidLogger

	rateLim            chan struct{}
	evaluateQueuedWork chan struct{}

	cancelLock  sync.Mutex
	cancelChans map[ffs.JobID]chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

// New returns a new instance of Scheduler which uses JobStore as its backing repository for state,
// HotStorage for the hot layer, and ColdStorage for the cold layer.
func New(ds datastore.TxnDatastore, l ffs.CidLogger, hs ffs.HotStorage, cs ffs.ColdStorage) (*Scheduler, error) {
	js, err := jstore.New(txndstr.Wrap(ds, "jstore"))
	if err != nil {
		return nil, fmt.Errorf("loading scheduler jobstore: %s", err)
	}
	as := astore.New(txndstr.Wrap(ds, "astore"))
	ts, err := trackstore.New(txndstr.Wrap(ds, "tstore"))
	if err != nil {
		return nil, fmt.Errorf("loading scheduler trackstore: %s", err)
	}

	cis := cistore.New(txndstr.Wrap(ds, "cistore"))
	ctx, cancel := context.WithCancel(context.Background())
	sch := &Scheduler{
		cs:  cs,
		hs:  hs,
		js:  js,
		as:  as,
		ts:  ts,
		cis: cis,
		l:   l,

		rateLim:            make(chan struct{}, maxParallelExecutions),
		evaluateQueuedWork: make(chan struct{}, 1),
		cancelChans:        make(map[ffs.JobID]chan struct{}),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go sch.run()
	return sch, nil
}

// PushConfig queues the specified StorageConfig to be executed as a new Job. It returns
// the created JobID for further tracking of its state.
func (s *Scheduler) PushConfig(iid ffs.APIID, c cid.Cid, cfg ffs.StorageConfig) (ffs.JobID, error) {
	return s.push(iid, c, cfg, cid.Undef)
}

// PushReplace queues a new StorageConfig to be executed as a new Job, replacing an oldCid that will be
// untrack in the Scheduler (i.e: deal renewals, repairing).
func (s *Scheduler) PushReplace(iid ffs.APIID, c cid.Cid, cfg ffs.StorageConfig, oldCid cid.Cid) (ffs.JobID, error) {
	if !oldCid.Defined() {
		return ffs.EmptyJobID, fmt.Errorf("cid can't be undefined")
	}
	return s.push(iid, c, cfg, oldCid)
}

func (s *Scheduler) push(iid ffs.APIID, c cid.Cid, cfg ffs.StorageConfig, oldCid cid.Cid) (ffs.JobID, error) {
	if !c.Defined() {
		return ffs.EmptyJobID, fmt.Errorf("cid can't be undefined")
	}
	if iid == ffs.EmptyInstanceID {
		return ffs.EmptyJobID, fmt.Errorf("empty API ID")
	}
	if err := cfg.Validate(); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("validating storage config: %s", err)
	}
	jid := ffs.NewJobID()
	j := ffs.Job{
		ID:     jid,
		APIID:  iid,
		Cid:    c,
		Status: ffs.Queued,
	}

	if err := s.js.Enqueue(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("enqueuing job: %s", err)
	}
	ctx := context.WithValue(context.Background(), ffs.CtxKeyJid, jid)
	s.l.Log(ctx, c, "Pushing new configuration...")

	aa := astore.StorageAction{
		APIID:       iid,
		Cid:         c,
		Cfg:         cfg,
		ReplacedCid: oldCid,
	}
	if err := s.as.Put(j.ID, aa); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving action for job: %s", err)
	}

	if err := s.ts.Put(iid, c, cfg); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving repairable/renewable storage config: %s", err)
	}

	select {
	case s.evaluateQueuedWork <- struct{}{}:
	default:
	}

	s.l.Log(ctx, c, "Configuration saved successfully")
	return jid, nil
}

// Untrack untracks a Cid for renewal and repair background crons.
func (s *Scheduler) Untrack(c cid.Cid) error {
	if err := s.ts.Remove(c); err != nil {
		return fmt.Errorf("removing cid from action store: %s", err)
	}
	return nil
}

// GetCidInfo returns the current storage state of a Cid. Returns ErrNotFound
// if there isn't information for a Cid.
func (s *Scheduler) GetCidInfo(c cid.Cid) (ffs.CidInfo, error) {
	info, err := s.cis.Get(c)
	if err == cistore.ErrNotFound {
		return ffs.CidInfo{}, ErrNotFound
	}
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting CidInfo from store: %s", err)
	}
	return info, nil
}

// ImportCidInfo imports Cid information manually. That's to say, will be CidInfo
// which wasn't generated by executing a Job, but provided externally.
func (s *Scheduler) ImportCidInfo(ci ffs.CidInfo) error {
	_, err := s.cis.Get(ci.Cid)
	if err != nil && err != cistore.ErrNotFound {
		return fmt.Errorf("checking if cid info already exists: %s", err)
	}
	if err != cistore.ErrNotFound {
		return fmt.Errorf("there is cid information for the provided cid")
	}
	if err := s.cis.Put(ci); err != nil {
		return fmt.Errorf("importing cid information: %s", err)
	}
	return nil
}

// GetCidFromHot returns an io.Reader of the data from the hot layer.
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
		if err == jstore.ErrNotFound {
			return ffs.Job{}, ErrNotFound
		}
		return ffs.Job{}, fmt.Errorf("get Job from store: %s", err)
	}
	return j, nil
}

// WatchJobs returns a channel to listen to Job status changes from a specified
// API instance. It immediately pushes the current Job state to the channel.
func (s *Scheduler) WatchJobs(ctx context.Context, c chan<- ffs.Job, iid ffs.APIID) error {
	return s.js.Watch(ctx, c, iid)
}

// WatchLogs writes to a channel all new logs for Cids. The context should be
// canceled when wanting to stop receiving updates to the channel.
func (s *Scheduler) WatchLogs(ctx context.Context, c chan<- ffs.LogEntry) error {
	return s.l.Watch(ctx, c)
}

// GetLogs returns history logs of a Cid.
func (s *Scheduler) GetLogs(ctx context.Context, c cid.Cid) ([]ffs.LogEntry, error) {
	lgs, err := s.l.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting logs: %s", err)
	}
	return lgs, nil
}

// Cancel cancels an executing Job.
func (s *Scheduler) Cancel(jid ffs.JobID) error {
	s.cancelLock.Lock()
	defer s.cancelLock.Unlock()
	cancelChan, ok := s.cancelChans[jid]
	if !ok {
		return nil
	}
	// The main scheduler loop is responsible for
	// deleting cancelChan from the map.
	close(cancelChan)
	return nil
}

// Close terminates the scheduler.
func (s *Scheduler) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	s.cancel()
	<-s.finished
	if err := s.js.Close(); err != nil {
		return fmt.Errorf("closing jobstore: %s", err)
	}
	return nil
}

func (s *Scheduler) run() {
	defer close(s.finished)
	if err := s.resumeStartedDeals(); err != nil {
		log.Errorf("resuming started deals: %s", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// Timer for evaluating renewable storage configs.
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(RenewalEvalFrequency):
				log.Debug("running renewal checks...")
				s.execRenewCron(s.ctx)
				log.Debug("renewal cron done")
			}
		}
	}()

	// Timer for evaluatin repairable storage configs.
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(RepairEvalFrequency):
				log.Debug("running repair checks...")
				s.execRepairCron(s.ctx)
				log.Debug("repair cron done")
			}
		}
	}()

	// Loop for new pushed storage configs.
	for {
		select {
		case <-s.ctx.Done():
			log.Infof("terminating scheduler daemon")
			wg.Wait()
			log.Infof("scheduler daemon terminated")
			return
		case <-s.evaluateQueuedWork:
			log.Debug("evaluating Job queue execution...")
			s.executeQueuedJobs(s.ctx)
			log.Debug("evaluation Job queue execution...")
		}
	}
}

func (s *Scheduler) resumeStartedDeals() error {
	ejids := s.js.GetExecutingJobs()
	// No need for rate limit since the number of "Executing"
	// jobs is always rate-limited on creation.
	for _, jid := range ejids {
		if s.ctx.Err() != nil {
			break
		}
		j, err := s.js.Get(jid)
		if err != nil {
			return fmt.Errorf("getting resumed queued job: %s", err)
		}
		s.rateLim <- struct{}{}
		go func(j ffs.Job) {
			log.Infof("resuming job %s with cid %s", j.ID, j.Cid)
			// We re-execute the pipeline as if was dequeued.
			// Both hot and cold storage can detect resumed job execution.
			s.executeQueuedJob(j)

			<-s.rateLim
		}(j)
	}
	return nil
}

// execRepairCron gets all repairable storage configs and
// reschedule them as if they were pushed. The scheduler main executing logic
// does whatever work is necessary to satisfy the storage config, thus
// it has repairing semantics too. If no work is needed, this scheduled
// job would have no real work done.
func (s *Scheduler) execRepairCron(ctx context.Context) {
	cids, err := s.ts.GetRepairables()
	if err != nil {
		log.Errorf("getting repairable cid configs from store: %s", err)
	}
	for _, c := range cids {
		s.l.Log(ctx, c, "Scheduling deal repair evaluation...")
		jid, err := s.scheduleRenewRepairJob(ctx, c)
		if err != nil {
			s.l.Log(ctx, c, "Scheduling deal repair errored: %s", err)

		} else {
			s.l.Log(ctx, c, "Job %s was queued for repair evaluation.", jid)
		}
	}
}

// execRenewCron gets all renewable storage configs and
// reschedule them as if they were pushed. The scheduler main executing logic
// will do renewals if necessary.
func (s *Scheduler) execRenewCron(ctx context.Context) {
	cids, err := s.ts.GetRenewables()
	if err != nil {
		log.Errorf("getting repairable cid configs from store: %s", err)
	}
	for _, c := range cids {
		s.l.Log(ctx, c, "Scheduling deal renew evaluation...")
		jid, err := s.scheduleRenewRepairJob(ctx, c)
		if err != nil {
			s.l.Log(ctx, c, "Scheduling deal renewal errored: %s", err)
		} else {
			s.l.Log(ctx, c, "Job %s was queued for renew evaluation.", jid)
		}
	}
}

func (s *Scheduler) scheduleRenewRepairJob(ctx context.Context, c cid.Cid) (ffs.JobID, error) {
	sc, iid, err := s.ts.Get(c)
	if err != nil {
		return "", fmt.Errorf("getting latest storage config: %s", err)
	}
	jid, err := s.push(iid, c, sc, cid.Undef)
	if err != nil {
		return "", fmt.Errorf("scheduling repair job: %s", err)
	}
	return jid, nil
}

func (s *Scheduler) executeQueuedJobs(ctx context.Context) {
	var err error
	var j *ffs.Job

forLoop:
	for {
		select {
		case <-ctx.Done():
			break forLoop
		case s.rateLim <- struct{}{}:
			// We have a slot, try pushing queued things.
		default:
			// If the execution pipeline is full, we can't
			// add new things as Executing.
			break forLoop
		}

		j, err = s.js.Dequeue()
		if err != nil {
			log.Errorf("getting queued jobs: %s", err)
			<-s.rateLim
			return
		}
		if j == nil {
			// Make the slot available again.
			<-s.rateLim
			break
		}
		go func(j ffs.Job) {
			s.executeQueuedJob(j)

			<-s.rateLim

			// Signal that there's a new open slot, and
			// that other blocked Jobs with the same Cid
			// can be executed.
			select {
			case s.evaluateQueuedWork <- struct{}{}:
			default:
			}
		}(*j)
	}
}

func (s *Scheduler) executeQueuedJob(j ffs.Job) {
	cancelChan := make(chan struct{})
	// Create chan to allow Job cancellation.
	s.cancelLock.Lock()
	s.cancelChans[j.ID] = cancelChan
	s.cancelLock.Unlock()
	defer func() {
		s.cancelLock.Lock()
		delete(s.cancelChans, j.ID)
		s.cancelLock.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ffs.CtxKeyJid, j.ID))
	defer cancel()
	go func() {
		// If the user called Cancel to cancel Job execution,
		// we cancel the context to finish.
		<-cancelChan
		cancel()
	}()

	// Get
	a, err := s.as.Get(j.ID)
	if err != nil {
		log.Errorf("getting push config action data from store: %s", err)
		if err := s.js.Finalize(j.ID, ffs.Failed, err, nil); err != nil {
			log.Errorf("changing job to failed: %s", err)
		}
		s.l.Log(ctx, a.Cid, "Job %s couldn't start: %s.", j.ID, err)
		return
	}

	// Execute
	s.l.Log(ctx, a.Cid, "Executing job %s...", j.ID)
	info, dealErrors, err := s.execute(ctx, a, j)

	// Something bad-enough happened to make Job
	// execution fail.
	if err != nil {
		log.Errorf("executing job %s: %s", j.ID, err)
		if err := s.js.Finalize(j.ID, ffs.Failed, err, dealErrors); err != nil {
			log.Errorf("changing job to failed: %s", err)
		}
		s.l.Log(ctx, a.Cid, "Job %s execution failed: %s", j.ID, err)
		return
	}
	// Save whatever stored information was completely/partially
	// done in execution.
	if err := s.cis.Put(info); err != nil {
		log.Errorf("saving cid info to store: %s", err)
	}

	finalStatus := ffs.Success
	// Detect if user-cancelation was triggered
	select {
	case <-cancelChan:
		finalStatus = ffs.Canceled
	default:
	}

	// Finalize Job, saving any deals errors happened during execution.
	if err := s.js.Finalize(j.ID, finalStatus, nil, dealErrors); err != nil {
		log.Errorf("changing job to success: %s", err)
	}
	s.l.Log(ctx, a.Cid, "Job %s execution finished with status %s.", j.ID, ffs.JobStatusStr[finalStatus])
}

// execute executes a Job. If an error is returned, it means that the Job
// should be considered failed. If error is nil, it still can return []ffs.DealError
// since some deals failing isn't necessarily a fatal Job config execution.
func (s *Scheduler) execute(ctx context.Context, a astore.StorageAction, job ffs.Job) (ffs.CidInfo, []ffs.DealError, error) {
	ci, err := s.getRefreshedInfo(ctx, a.Cid)
	if err != nil {
		return ffs.CidInfo{}, nil, fmt.Errorf("getting current cid info from store: %s", err)
	}

	if a.ReplacedCid.Defined() {
		if err := s.Untrack(a.ReplacedCid); err != nil && err != astore.ErrNotFound {
			return ffs.CidInfo{}, nil, fmt.Errorf("untracking replaced cid: %s", err)
		}
	}

	s.l.Log(ctx, a.Cid, "Ensuring Hot-Storage satisfies the configuration...")
	hot, err := s.executeHotStorage(ctx, ci, a.Cfg.Hot, a.Cfg.Cold.Filecoin.Addr, a.ReplacedCid)
	if err != nil {
		s.l.Log(ctx, a.Cid, "Hot-Storage excution failed.")
		return ffs.CidInfo{}, nil, fmt.Errorf("executing hot-storage config: %s", err)
	}
	s.l.Log(ctx, a.Cid, "Hot-Storage execution ran successfully.")

	s.l.Log(ctx, a.Cid, "Ensuring Cold-Storage satisfies the configuration...")
	cold, errors, err := s.executeColdStorage(ctx, ci, a.Cfg.Cold)
	if err != nil {
		s.l.Log(ctx, a.Cid, "Cold-Storage execution failed.")
		return ffs.CidInfo{}, errors, fmt.Errorf("executing cold-storage config: %s", err)
	}
	s.l.Log(ctx, a.Cid, "Cold-Storage execution ran successfully.")

	return ffs.CidInfo{
		JobID:   job.ID,
		Cid:     a.Cid,
		Hot:     hot,
		Cold:    cold,
		Created: time.Now(),
	}, errors, nil
}

func (s *Scheduler) executeHotStorage(ctx context.Context, curr ffs.CidInfo, cfg ffs.HotConfig, waddr string, replaceCid cid.Cid) (ffs.HotInfo, error) {
	if cfg.Enabled == curr.Hot.Enabled {
		s.l.Log(ctx, curr.Cid, "No actions needed in Hot Storage.")
		return curr.Hot, nil
	}

	if !cfg.Enabled {
		if err := s.hs.Remove(ctx, curr.Cid); err != nil {
			return ffs.HotInfo{}, fmt.Errorf("removing from hot storage: %s", err)
		}
		s.l.Log(ctx, curr.Cid, "Cid successfully removed from Hot Storage.")
		return ffs.HotInfo{Enabled: false}, nil
	}

	sctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.Ipfs.AddTimeout))
	defer cancel()

	var size int
	var err error
	if !replaceCid.Defined() {
		size, err = s.hs.Store(sctx, curr.Cid)
	} else {
		s.l.Log(ctx, curr.Cid, "Replace of previous pin %s", replaceCid)
		size, err = s.hs.Replace(sctx, replaceCid, curr.Cid)
	}
	if err != nil {
		s.l.Log(ctx, curr.Cid, "Direct fetching from IPFS wasn't possible.")
		if !cfg.AllowUnfreeze || len(curr.Cold.Filecoin.Proposals) == 0 {
			s.l.Log(ctx, curr.Cid, "Unfreeze is disabled or active Filecoin deals are unavailable.")
			return ffs.HotInfo{}, fmt.Errorf("pinning cid in hot storage: %s", err)
		}
		s.l.Log(ctx, curr.Cid, "Unfreezing from Filecoin...")
		if len(curr.Cold.Filecoin.Proposals) == 0 {
			return ffs.HotInfo{}, fmt.Errorf("no active deals to make retrieval possible")
		}
		var pieceCid *cid.Cid
		for _, p := range curr.Cold.Filecoin.Proposals {
			if p.PieceCid != cid.Undef {
				pieceCid = &p.PieceCid
				break
			}
		}

		if err := s.cs.Fetch(ctx, curr.Cold.Filecoin.DataCid, pieceCid, waddr); err != nil {
			return ffs.HotInfo{}, fmt.Errorf("unfreezing from Cold Storage: %s", err)
		}
		s.l.Log(ctx, curr.Cid, "Unfrozen successfully, saving in Hot-Storage...")
		size, err = s.hs.Store(ctx, curr.Cold.Filecoin.DataCid)
		if err != nil {
			return ffs.HotInfo{}, fmt.Errorf("pinning unfrozen cid: %s", err)
		}
	}
	return ffs.HotInfo{
		Enabled: true,
		Size:    size,
		Ipfs: ffs.IpfsHotInfo{
			Created: time.Now(),
		},
	}, nil
}

func (s *Scheduler) getRefreshedInfo(ctx context.Context, c cid.Cid) (ffs.CidInfo, error) {
	var err error
	ci, err := s.cis.Get(c)
	if err != nil {
		if err != cistore.ErrNotFound {
			return ffs.CidInfo{}, ErrNotFound
		}
		return ffs.CidInfo{Cid: c}, nil // Default value has both storages disabled
	}

	ci.Hot, err = s.getRefreshedHotInfo(ctx, c, ci.Hot)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting refreshed hot info: %s", err)
	}

	ci.Cold, err = s.getRefreshedColdInfo(ctx, ci.Cold)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting refreshed cold info: %s", err)
	}

	return ci, nil
}

func (s *Scheduler) getRefreshedHotInfo(ctx context.Context, c cid.Cid, curr ffs.HotInfo) (ffs.HotInfo, error) {
	var err error
	curr.Enabled, err = s.hs.IsStored(ctx, c)
	if err != nil {
		return ffs.HotInfo{}, err
	}
	return curr, nil
}

func (s *Scheduler) getRefreshedColdInfo(ctx context.Context, curr ffs.ColdInfo) (ffs.ColdInfo, error) {
	var err error
	activeDeals := make([]ffs.FilStorage, 0, len(curr.Filecoin.Proposals))
	for _, fp := range curr.Filecoin.Proposals {
		active := true
		// Consider the border-case of imported deals which
		// didn't provide the ProposalCid of the deal.
		if fp.ProposalCid != cid.Undef {
			active, err = s.cs.IsFilDealActive(ctx, fp.ProposalCid)
			if err != nil {
				return ffs.ColdInfo{}, fmt.Errorf("getting deal state of proposal %s: %s", fp.ProposalCid, err)
			}
		}
		if active {
			activeDeals = append(activeDeals, fp)
		}
	}
	curr.Filecoin.Proposals = activeDeals
	return curr, nil
}

func (s *Scheduler) executeColdStorage(ctx context.Context, curr ffs.CidInfo, cfg ffs.ColdConfig) (ffs.ColdInfo, []ffs.DealError, error) {
	if !cfg.Enabled {
		s.l.Log(ctx, curr.Cid, "Cold-Storage was disabled, Filecoin deals will eventually expire.")
		return curr.Cold, nil, nil
	}

	// 1. If we recognize there were some unfinished started deals, then
	// Powergate was closed while that was being executed. If that's the case
	// we resume tracking those deals until they finish.
	sds, err := s.js.GetStartedDeals(curr.Cid)
	if err != nil {
		return ffs.ColdInfo{}, nil, fmt.Errorf("checking for started deals: %s", err)
	}
	var allErrors []ffs.DealError
	if len(sds) > 0 {
		s.l.Log(ctx, curr.Cid, "Resuming %d dettached executing deals...", len(sds))
		okResumedDeals, failedResumedDeals := s.waitForDeals(ctx, curr.Cid, sds)
		s.l.Log(ctx, curr.Cid, "A total of %d resumed deals finished successfully", len(okResumedDeals))
		allErrors = append(allErrors, failedResumedDeals...)
		// Append the resumed and confirmed deals to the current active proposals
		curr.Cold.Filecoin.Proposals = append(okResumedDeals, curr.Cold.Filecoin.Proposals...)
	}

	// 2. If this Storage Config is renewable, then let's check if any of the existing deals
	// should be renewed, and do it.
	if cfg.Filecoin.Renew.Enabled {
		if curr.Hot.Enabled {
			s.l.Log(ctx, curr.Cid, "Checking deal renewals...")
			newFilInfo, errors, err := s.cs.EnsureRenewals(ctx, curr.Cid, curr.Cold.Filecoin, cfg.Filecoin)
			if err != nil {
				s.l.Log(ctx, curr.Cid, "Deal renewal process couldn't be executed: %s", err)
			} else {
				for _, e := range errors {
					s.l.Log(ctx, curr.Cid, "Deal deal renewal errored. ProposalCid: %s, Miner: %s, Cause: %s", e.ProposalCid, e.Miner, e.Message)
				}
				numDeals := len(newFilInfo.Proposals) - len(curr.Cold.Filecoin.Proposals)
				if numDeals > 0 {
					// If renew process created deals, we eagerly save this information in the datastore.
					// Further work about the new storage config could decide the Job failed and we'd lose
					// this information if not saved.
					if err := s.cis.Put(curr); err != nil {
						return ffs.ColdInfo{}, nil, fmt.Errorf("eager saving of new info: %s", err)
					}
					s.l.Log(ctx, curr.Cid, "A total of %d new deals were created in the renewal process", numDeals)
				}
				s.l.Log(ctx, curr.Cid, "Deal renewal evaluated successfully")
				curr.Cold.Filecoin = newFilInfo

				if err := s.cis.Put(curr); err != nil {
					log.Errorf("saving cid info to store: %s", err)
				}
			}
		} else {
			// (**) Renewable note:
			// This shouldn't happen since it would be an invalid Storage Config.
			// We can only accept "Repair" if Hot Storage is enabled.
			// We can the feature to retrieve the data from a miner,
			// put it in Hot Storage, and then try the renewal. Looks to me
			// we should be sure about doing that since it would be paying
			// for retrieval to later discard the data. Sounds like Filecoin
			// should allow renewing a deal without the need of sending the data
			// again. Still not clear.
			return ffs.ColdInfo{}, nil, fmt.Errorf("invalid storage configuration, can't be renewable with disabled hot storage")
		}
	}

	// 3. Now that we have final information about what deals are really active,
	// we calculate how many new deals should be made to ensure the desired RepFactor.
	// If the current active deals is equal or greater than desired, do nothing. If not, make
	// whatever extra deals we need to make that true.

	// Do we need to do some work?
	if cfg.Filecoin.RepFactor-len(curr.Cold.Filecoin.Proposals) <= 0 {
		s.l.Log(ctx, curr.Cid, "The current replication factor is equal or higher than desired, avoiding making new deals.")
		return curr.Cold, nil, nil
	}

	// The answer is yes, calculate how many extra deals we need and create them.
	deltaFilConfig := createDeltaFilConfig(cfg, curr.Cold.Filecoin)
	s.l.Log(ctx, curr.Cid, "Current replication factor is lower than desired, making %d new deals...", deltaFilConfig.RepFactor)
	startedProposals, rejectedProposals, size, err := s.cs.Store(ctx, curr.Cid, deltaFilConfig)
	if err != nil {
		s.l.Log(ctx, curr.Cid, "Starting deals failed, with cause: %s", err)
		return ffs.ColdInfo{}, rejectedProposals, err
	}
	allErrors = append(allErrors, rejectedProposals...)

	// If *none* of the tried proposals succeeded, then the Job fails.
	if len(startedProposals) == 0 {
		return ffs.ColdInfo{}, allErrors, fmt.Errorf("all proposals were rejected")
	}

	// Track all deals that weren't rejected, just in case Powergate crashes/closes before
	// we see them finalize, so they can be detected and resumed on starting Powergate again (point 1. above)
	if err := s.js.AddStartedDeals(curr.Cid, startedProposals); err != nil {
		return ffs.ColdInfo{}, rejectedProposals, err
	}

	// Wait for started deals.
	okDeals, failedDeals := s.waitForDeals(ctx, curr.Cid, startedProposals)
	allErrors = append(allErrors, failedDeals...)
	if err := s.js.RemoveStartedDeals(curr.Cid); err != nil {
		return ffs.ColdInfo{}, allErrors, fmt.Errorf("removing temporal started deals storage: %s", err)
	}

	// If the Job wasn't canceled, and not even one deal finished succcessfully,
	// consider this Job execution a failure.
	if ctx.Err() == nil && len(failedDeals) == len(startedProposals) {
		return ffs.ColdInfo{}, allErrors, fmt.Errorf("all started deals failed")
	}

	// At least 1 of the proposal deals reached a successful final status, Job succeeds.
	return ffs.ColdInfo{
		Enabled: true,
		Filecoin: ffs.FilInfo{
			DataCid:   curr.Cid,
			Size:      size,
			Proposals: append(okDeals, curr.Cold.Filecoin.Proposals...), // Append to any existing other proposals
		},
	}, allErrors, nil
}

func (s *Scheduler) waitForDeals(ctx context.Context, c cid.Cid, startedProposals []cid.Cid) ([]ffs.FilStorage, []ffs.DealError) {
	s.l.Log(ctx, c, "Watching deals unfold...")

	var failedDeals []ffs.DealError
	var okDeals []ffs.FilStorage
	var wg sync.WaitGroup
	var lock sync.Mutex
	wg.Add(len(startedProposals))
	for _, pc := range startedProposals {
		pc := pc
		go func() {
			defer wg.Done()

			res, err := s.cs.WaitForDeal(ctx, c, pc)
			var dealError ffs.DealError
			if err != nil {
				if !errors.As(err, &dealError) {
					dealError = ffs.DealError{
						ProposalCid: pc,
						Message:     fmt.Sprintf("waiting for deal finality: %s", err),
					}
				}
				lock.Lock()
				failedDeals = append(failedDeals, dealError)
				lock.Unlock()
				return
			}
			lock.Lock()
			okDeals = append(okDeals, res)
			lock.Unlock()
		}()
	}
	wg.Wait()
	return okDeals, failedDeals
}

func createDeltaFilConfig(cfg ffs.ColdConfig, curr ffs.FilInfo) ffs.FilConfig {
	res := cfg.Filecoin
	res.RepFactor = cfg.Filecoin.RepFactor - len(curr.Proposals)
	for _, p := range curr.Proposals {
		res.ExcludedMiners = append(res.ExcludedMiners, p.Miner)
	}
	return res
}

func (s *Scheduler) StartRetrieval(iid ffs.APIID, rid ffs.RetrievalID, pyCid, piCid cid.Cid, sel string, miners []string, walletAddr string, maxPrice uint64) (ffs.JobID, error) {

}
