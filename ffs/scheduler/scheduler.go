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
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/astore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/cistore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/ristore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/rjstore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/sjstore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/trackstore"
	"github.com/textileio/powergate/v2/notifications"
	txndstr "github.com/textileio/powergate/v2/txndstransform"
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

// Scheduler receives actions to store a Cid in hot and cold storage. These actions are
// created as Jobs which have a lifecycle that can be watched by external actors.
// This Jobs are executed by delegating the work to the hot and cold storage configured for
// the scheduler.
type Scheduler struct {
	cs       ffs.ColdStorage
	hs       ffs.HotStorage
	sjs      *sjstore.Store
	rjs      *rjstore.Store
	as       *astore.Store
	ts       *trackstore.Store
	cis      *cistore.Store
	ris      *ristore.Store
	l        ffs.JobLogger
	notifier notifications.Notifier

	sr2RepFactor        func() (int, error)
	dealFinalityTimeout time.Duration

	gcLock sync.Mutex
	gc     GCConfig

	sd         storageDaemon
	rd         retrievalDaemon
	cancelLock sync.Mutex
	jobsCancel map[ffs.JobID]chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

// storageDaemon contains components used by
// the daemon attending the storage job executions.
type storageDaemon struct {
	rateLim       chan struct{}
	evaluateQueue chan struct{}
}

// retrievalDaemon contains components used by
// the daemon attending the retrieval jobs executions.
type retrievalDaemon struct {
	rateLim       chan struct{}
	evaluateQueue chan struct{}
}

// GCConfig provides configuration for FFS GC.
type GCConfig struct {
	StageGracePeriod time.Duration
	AutoGCInterval   time.Duration
}

// New returns a new instance of Scheduler which uses JobStore as its backing repository for state,
// HotStorage for the hot layer, and ColdStorage for the cold layer.
func New(ds datastore.TxnDatastore, l ffs.JobLogger, hs ffs.HotStorage, cs ffs.ColdStorage, maxParallel int, dealFinalityTimeout time.Duration, sr2rf func() (int, error), gcConfig GCConfig, notifier notifications.Notifier) (*Scheduler, error) {
	sjs, err := sjstore.New(txndstr.Wrap(ds, "sjstore"), notifier)
	if err != nil {
		return nil, fmt.Errorf("loading stroage jobstore: %s", err)
	}
	rjs, err := rjstore.New(txndstr.Wrap(ds, "rjstore"))
	if err != nil {
		return nil, fmt.Errorf("loading retrieval jobstore: %s", err)
	}
	as := astore.New(txndstr.Wrap(ds, "astore"))
	ts, err := trackstore.New(txndstr.Wrap(ds, "tstore"))
	if err != nil {
		return nil, fmt.Errorf("loading scheduler trackstore: %s", err)
	}

	cis := cistore.New(txndstr.Wrap(ds, "cistore_v2"))
	ris := ristore.New(txndstr.Wrap(ds, "ristore"))

	ctx, cancel := context.WithCancel(context.Background())
	sch := &Scheduler{
		cs: cs,
		hs: hs,

		sjs: sjs,
		rjs: rjs,

		as: as,
		ts: ts,

		cis: cis,
		ris: ris,

		l:  l,
		gc: gcConfig,

		notifier: notifier,

		jobsCancel: make(map[ffs.JobID]chan struct{}),
		sd: storageDaemon{
			rateLim:       make(chan struct{}, maxParallel),
			evaluateQueue: make(chan struct{}, 1),
		},
		rd: retrievalDaemon{
			rateLim:       make(chan struct{}, maxParallel),
			evaluateQueue: make(chan struct{}, 1),
		},

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),

		sr2RepFactor:        sr2rf,
		dealFinalityTimeout: dealFinalityTimeout,
	}

	go sch.run()

	return sch, nil
}

// GetCidFromHot returns an io.Reader of the data from hot storage.
func (s *Scheduler) GetCidFromHot(ctx context.Context, c cid.Cid) (io.Reader, error) {
	r, err := s.hs.Get(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting %s from hot storage: %s", c, err)
	}
	return r, nil
}

// Cancel cancels an executing Job.
func (s *Scheduler) Cancel(jid ffs.JobID) error {
	s.cancelLock.Lock()
	defer s.cancelLock.Unlock()

	ok, err := s.sjs.CancelQueued(jid)
	if err != nil {
		return fmt.Errorf("canceling potentially queued job: %s", err)
	}
	// The Job was Queued, and switched to Cancel; we're done.
	if ok {
		return nil
	}

	cancelChan, ok := s.jobsCancel[jid]
	// The Job isn't executing; nothing to do.
	if !ok {
		return nil
	}

	// Send cancel signal to executing job.
	// The main scheduler loop is responsible for
	// deleting cancelChan from the map.
	select {
	case cancelChan <- struct{}{}:
	default:
		// If this channel was already signaled,
		// don't block, this is just a cancel retry.
	}
	return nil
}

// GCStaged runs a unpinned garbage collection of stage-pins.
func (s *Scheduler) GCStaged(ctx context.Context) ([]cid.Cid, error) {
	return s.gcStaged(ctx, 0)
}

// PinnedCids returns the pinned cids from Hot-Storage.
func (s *Scheduler) PinnedCids(ctx context.Context) ([]ffs.PinnedCid, error) {
	res, err := s.hs.PinnedCids(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting pinned cids from hot-storage: %s", err)
	}
	return res, nil
}

func (s *Scheduler) gcStaged(ctx context.Context, gracePeriod time.Duration) ([]cid.Cid, error) {
	s.gcLock.Lock()
	defer s.gcLock.Unlock()
	log.Infof("running scheduler gc...")
	cids := map[cid.Cid]struct{}{}

	qj, _, _, err := s.sjs.List(sjstore.ListConfig{Select: sjstore.Queued})
	if err != nil {
		return nil, fmt.Errorf("listing queued storage jobs: %s", err)
	}
	for _, j := range qj {
		cids[j.Cid] = struct{}{}
	}
	ej, _, _, err := s.sjs.List(sjstore.ListConfig{Select: sjstore.Executing})
	if err != nil {
		return nil, fmt.Errorf("listing executing storage jobs: %s", err)
	}
	for _, j := range ej {
		cids[j.Cid] = struct{}{}
	}

	excludedCids := make([]cid.Cid, 0, len(cids))
	for c := range cids {
		excludedCids = append(excludedCids, c)
	}

	gced, err := s.hs.GCStaged(ctx, excludedCids, time.Now().Add(-gracePeriod))
	if err != nil {
		return nil, fmt.Errorf("hot-storage gc: %s", err)
	}

	log.Infof("scheduler gc ran with %d excluded cids, unpinning %d staged cids", len(excludedCids), len(gced))

	return gced, nil
}

// Close terminates the scheduler.
func (s *Scheduler) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
	s.cancel()
	<-s.finished
	if err := s.sjs.Close(); err != nil {
		return fmt.Errorf("closing jobstore: %s", err)
	}
	return nil
}

// run spins the long-running goroutines that will execute
// queued storage, retrieval jobs, renewals, repairs, and gc.
func (s *Scheduler) run() {
	defer close(s.finished)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		if s.gc.AutoGCInterval == 0 {
			return
		}

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(s.gc.AutoGCInterval):
				if _, err := s.gcStaged(s.ctx, s.gc.StageGracePeriod); err != nil {
					log.Errorf("automatic gc: %s", err)
				}
			}
		}
	}()

	if err := s.resumeStartedDeals(); err != nil {
		log.Errorf("resuming started deals: %s", err)
		return
	}

	// Timer for evaluating renewable storage configs.
	wg.Add(1)
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
	wg.Add(1)
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

	// Loop for retrievals jobs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.rd.evaluateQueue:
				log.Debug("evaluating retrieval job queue...")
				s.execQueuedRetrievals(s.ctx)
				log.Debug("retrieval job queue evaluated")
			}
		}
	}()

	// Force a first evaluation on start.
	go func() { s.sd.evaluateQueue <- struct{}{} }()

	// Loop for new pushed storage configs.
	for {
		select {
		case <-s.ctx.Done():
			log.Infof("terminating scheduler daemon")
			wg.Wait()
			log.Infof("scheduler daemon terminated")
			return
		case <-s.sd.evaluateQueue:
			s.printStats()
			s.execQueuedStorages(s.ctx)
			s.printStats()
		}
	}
}

func (s *Scheduler) printStats() {
	stats := s.sjs.GetStats()
	log.Infof("storage job total queued: %d, total executing: %d", stats.TotalQueued, stats.TotalExecuting)
}

func (s *Scheduler) resumeStartedDeals() error {
	ejids := s.sjs.GetExecutingJobIDs()
	// No need for rate limit since the number of "Executing"
	// jobs is always rate-limited on creation.
	for _, jid := range ejids {
		if s.ctx.Err() != nil {
			break
		}
		j, err := s.sjs.Get(jid)
		if err != nil {
			return fmt.Errorf("getting resumed queued job: %s", err)
		}

		s.printStats()
		s.sd.rateLim <- struct{}{}
		go func(j ffs.StorageJob) {
			log.Infof("resuming job %s with cid %s", j.ID, j.Cid)
			// We re-execute the pipeline as if was dequeued.
			// Both hot and cold storage can detect resumed job execution.
			s.executeQueuedStorage(j)

			s.printStats()
			<-s.sd.rateLim

			// Signal that a free slot is available for a queued job.
			select {
			case s.sd.evaluateQueue <- struct{}{}:
			default:
			}
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
	tcids, err := s.ts.GetRepairables()
	if err != nil {
		log.Errorf("getting repairable cid configs from store: %s", err)
		return
	}
	for _, tc := range tcids {
		for _, sc := range tc.Tracked {
			lCtx := context.WithValue(ctx, ffs.CtxStorageCid, tc.Cid)
			lCtx = context.WithValue(lCtx, ffs.CtxAPIID, sc.IID)
			s.l.Log(lCtx, "Scheduling deal repair evaluation...")
			jid, err := s.push(sc.IID, tc.Cid, sc.StorageConfig, cid.Undef)
			if err != nil {
				s.l.Log(lCtx, "Scheduling deal repair errored: %s", err)
			} else {
				s.l.Log(lCtx, "Job %s was queued for repair evaluation.", jid)
			}
		}
	}
}

// execRenewCron gets all renewable storage configs and
// reschedule them as if they were pushed. The scheduler main executing logic
// will do renewals if necessary.
func (s *Scheduler) execRenewCron(ctx context.Context) {
	tcids, err := s.ts.GetRenewables()
	if err != nil {
		log.Errorf("getting repairable cid configs from store: %s", err)
	}
	for _, tc := range tcids {
		for _, sc := range tc.Tracked {
			lCtx := context.WithValue(ctx, ffs.CtxStorageCid, tc.Cid)
			lCtx = context.WithValue(lCtx, ffs.CtxAPIID, sc.IID)
			s.l.Log(lCtx, "Scheduling deal renew evaluation...")
			jid, err := s.push(sc.IID, tc.Cid, sc.StorageConfig, cid.Undef)
			if err != nil {
				s.l.Log(lCtx, "Scheduling deal renewal errored: %s", err)
			} else {
				s.l.Log(lCtx, "Job %s was queued for renew evaluation.", jid)
			}
		}
	}
}

func (s *Scheduler) execQueuedStorages(ctx context.Context) {
	var err error
	var j *ffs.StorageJob

forLoop:
	for {
		select {
		case <-ctx.Done():
			break forLoop
		case s.sd.rateLim <- struct{}{}:
			// We have a slot, try pushing queued things.
		default:
			// If the execution pipeline is full, we can't
			// add new things as Executing.
			break forLoop
		}

		j, err = s.sjs.Dequeue(ffs.EmptyInstanceID)
		if err != nil {
			log.Errorf("getting queued jobs: %s", err)
			<-s.sd.rateLim
			return
		}
		if j == nil {
			// Make the slot available again.
			<-s.sd.rateLim
			break
		}
		go func(j ffs.StorageJob) {
			s.executeQueuedStorage(j)
			s.printStats()
			<-s.sd.rateLim

			// Signal that there's a new open slot, and
			// that other blocked Jobs with the same Cid
			// can be executed.
			select {
			case s.sd.evaluateQueue <- struct{}{}:
			default:
			}
		}(*j)
	}
}

func (s *Scheduler) executeQueuedStorage(j ffs.StorageJob) {
	cancelChan := make(chan struct{}, 1)
	// Create chan to allow Job cancellation.
	s.cancelLock.Lock()
	s.jobsCancel[j.ID] = cancelChan
	s.cancelLock.Unlock()
	defer func() {
		s.cancelLock.Lock()
		delete(s.jobsCancel, j.ID)
		s.cancelLock.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ffs.CtxKeyJid, j.ID))
	defer cancel()
	ctx = context.WithValue(ctx, ffs.CtxStorageCid, j.Cid)
	ctx = context.WithValue(ctx, ffs.CtxAPIID, j.APIID)

	var cancelLock sync.Mutex
	var canceled bool
	go func() {
		// If the user called Cancel to cancel Job execution,
		// we cancel the context to finish.
		<-cancelChan
		cancelLock.Lock()
		canceled = true
		cancelLock.Unlock()
		cancel()
	}()

	// Get
	a, err := s.as.GetStorageAction(j.ID)
	if err != nil {
		log.Errorf("getting push config action data from store: %s", err)
		if err := s.sjs.Finalize(j.ID, ffs.Failed, err, nil); err != nil {
			log.Errorf("changing job to failed: %s", err)
		}
		s.l.Log(ctx, "Job %s couldn't start: %s.", j.ID, err)
		return
	}

	s.notifier.RegisterStorageJob(j, a.Cfg.Notifications)

	// Execute
	s.l.Log(ctx, "Executing job %s...", j.ID)
	dealUpdates := s.sjs.MonitorJob(j)
	info, dealErrors, err := s.executeStorage(ctx, a, j, dealUpdates)
	close(dealUpdates)
	// Something bad-enough happened to make Job
	// execution fail.
	if err != nil {
		log.Errorf("executing job %s: %s", j.ID, err)
		if err := s.sjs.Finalize(j.ID, ffs.Failed, err, dealErrors); err != nil {
			log.Errorf("changing job to failed: %s", err)
		}
		s.l.Log(ctx, "Job %s execution failed: %s", j.ID, err)
		return
	}
	// Save whatever stored information was completely/partially
	// done in execution.
	if err := s.cis.Put(info); err != nil {
		log.Errorf("saving cid info to store: %s", err)
	}

	finalStatus := ffs.Success
	// Detect if user-cancelation was triggered
	cancelLock.Lock()
	if canceled {
		finalStatus = ffs.Canceled
	}
	cancelLock.Unlock()

	// Finalize Job, saving any deals errors happened during execution.
	if err := s.sjs.Finalize(j.ID, finalStatus, nil, dealErrors); err != nil {
		log.Errorf("changing job to success: %s", err)
	}

	s.l.Log(ctx, "Job %s execution finished with status %s.", j.ID, ffs.JobStatusStr[finalStatus])
}

func (s *Scheduler) execQueuedRetrievals(ctx context.Context) {
	var err error
	var j *ffs.RetrievalJob

forLoop:
	for {
		select {
		case <-ctx.Done():
			break forLoop
		case s.rd.rateLim <- struct{}{}:
			// We have a slot, try pushing queued things.
		default:
			// If the execution pipeline is full, we can't
			// add new things as Executing.
			break forLoop
		}

		j, err = s.rjs.Dequeue()
		if err != nil {
			log.Errorf("getting queued jobs: %s", err)
			<-s.rd.rateLim
			return
		}
		if j == nil {
			// Make the slot available again.
			<-s.rd.rateLim
			break
		}
		go func(j ffs.RetrievalJob) {
			s.executeQueuedRetrievals(j)

			<-s.rd.rateLim

			// Signal that there's a new open slot, and
			// that other blocked Jobs with the same Cid
			// can be executed.
			select {
			case s.rd.evaluateQueue <- struct{}{}:
			default:
			}
		}(*j)
	}
}

func (s *Scheduler) executeQueuedRetrievals(j ffs.RetrievalJob) {
	cancelChan := make(chan struct{})
	// Create chan to allow Job cancellation.
	s.cancelLock.Lock()
	s.jobsCancel[j.ID] = cancelChan
	s.cancelLock.Unlock()
	defer func() {
		s.cancelLock.Lock()
		delete(s.jobsCancel, j.ID)
		s.cancelLock.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), ffs.CtxKeyJid, j.ID))
	defer cancel()
	ctx = context.WithValue(ctx, ffs.CtxRetrievalID, j.RetrievalID)
	go func() {
		// If the user called Cancel to cancel Job execution,
		// we cancel the context to finish.
		<-cancelChan
		cancel()
	}()

	// Get
	a, err := s.as.GetRetrievalAction(j.ID)
	if err != nil {
		log.Errorf("getting job action data from store: %s", err)
		if err := s.rjs.Finalize(j.ID, ffs.Failed, err); err != nil {
			log.Errorf("changing job to failed: %s", err)
		}
		s.l.Log(ctx, "Job %s couldn't start: %s.", j.ID, err)
		return
	}

	// Execute
	s.l.Log(ctx, "Executing job %s...", j.ID)
	info, err := s.executeRetrieval(ctx, a, j)

	// Something bad-enough happened to make Job
	// execution fail.
	if err != nil {
		log.Errorf("executing retrieval job %s: %s", j.ID, err)
		if err := s.rjs.Finalize(j.ID, ffs.Failed, err); err != nil {
			log.Errorf("changing retrieval job status to failed: %s", err)
		}
		s.l.Log(ctx, "Job %s execution failed: %s", j.ID, err)
		return
	}
	// Save whatever stored information was completely/partially
	// done in execution.
	if err := s.ris.Put(info); err != nil {
		log.Errorf("saving retrieval info into store: %s", err)
	}

	finalStatus := ffs.Success
	// Detect if user-cancelation was triggered
	select {
	case <-cancelChan:
		finalStatus = ffs.Canceled
	default:
	}

	// Finalize Job, saving any deals errors happened during execution.
	if err := s.rjs.Finalize(j.ID, finalStatus, nil); err != nil {
		log.Errorf("changing retrieval job to success: %s", err)
	}
	s.l.Log(ctx, "Retrieval job %s execution finished with status %s.", j.ID, ffs.JobStatusStr[finalStatus])
}
