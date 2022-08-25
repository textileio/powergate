package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/astore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/cistore"
	"github.com/textileio/powergate/v2/ffs/scheduler/internal/sjstore"
	"github.com/textileio/powergate/v2/notifications"
)

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
	j := ffs.StorageJob{
		ID:        jid,
		APIID:     iid,
		Cid:       c,
		Status:    ffs.Queued,
		CreatedAt: time.Now().Unix(),
	}

	ctx := context.WithValue(context.Background(), ffs.CtxKeyJid, jid)
	ctx = context.WithValue(ctx, ffs.CtxStorageCid, c)
	ctx = context.WithValue(ctx, ffs.CtxAPIID, iid)
	s.l.Log(ctx, "Pushing new configuration...")

	aa := astore.StorageAction{
		APIID:       iid,
		Cid:         c,
		Cfg:         cfg,
		ReplacedCid: oldCid,
	}
	if err := s.as.PutStorageAction(j.ID, aa); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving action for job: %s", err)
	}

	if err := s.ts.Put(iid, c, cfg); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving repairable/renewable storage config: %s", err)
	}

	if err := s.sjs.Enqueue(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("enqueuing job: %s", err)
	}
	if jid := s.sjs.GetExecutingJob(iid, c); jid != nil {
		s.l.Log(ctx, "Job %s is already being executed for the same data, this job will be queued until it finishes or is canceled.", jid)
	}

	select {
	case s.sd.evaluateQueue <- struct{}{}:
	default:
	}

	s.l.Log(ctx, "Configuration saved successfully")

	s.notifier.RegisterJob(j.ID, cfg.Notifications)
	s.notifier.Alert(notifications.DiskSpaceAlert{JobID: jid}, cfg.Notifications)

	return jid, nil
}

// Untrack untracks a Cid from iid for renewal and repair background crons.
func (s *Scheduler) Untrack(iid ffs.APIID, c cid.Cid) error {
	if err := s.ts.Remove(iid, c); err != nil {
		return fmt.Errorf("removing cid from action store: %s", err)
	}
	return nil
}

// GetStorageInfo returns the current storage state of a Cid for a APIID. Returns ErrNotFound
// if there isn't information for a Cid.
func (s *Scheduler) GetStorageInfo(iid ffs.APIID, c cid.Cid) (ffs.StorageInfo, error) {
	info, err := s.cis.Get(iid, c)
	if err == cistore.ErrNotFound {
		return ffs.StorageInfo{}, ErrNotFound
	}
	if err != nil {
		return ffs.StorageInfo{}, fmt.Errorf("getting StorageInfo from store: %s", err)
	}
	return info, nil
}

// ListStorageInfo returns a list of StorageInfo matching any provided query options.
func (s *Scheduler) ListStorageInfo(iids []ffs.APIID, cids []cid.Cid) ([]ffs.StorageInfo, error) {
	res, err := s.cis.List(iids, cids)
	if err != nil {
		return nil, fmt.Errorf("listing storage info from cistore: %v", err)
	}
	return res, nil
}

// StorageJob the current storage state of a Job.
func (s *Scheduler) StorageJob(jid ffs.JobID) (ffs.StorageJob, error) {
	j, err := s.sjs.Get(jid)
	if err != nil {
		if err == sjstore.ErrNotFound {
			return ffs.StorageJob{}, ErrNotFound
		}
		return ffs.StorageJob{}, fmt.Errorf("get Job from store: %s", err)
	}
	return j, nil
}

// WatchJobs returns a channel to listen to Job status changes from a specified
// API instance. It immediately pushes the current Job state to the channel.
func (s *Scheduler) WatchJobs(ctx context.Context, c chan<- ffs.StorageJob, iid ffs.APIID) error {
	return s.sjs.Watch(ctx, c, iid)
}

// WatchLogs writes to a channel all new logs for Cids. The context should be
// canceled when wanting to stop receiving updates to the channel.
func (s *Scheduler) WatchLogs(ctx context.Context, c chan<- ffs.LogEntry) error {
	return s.l.Watch(ctx, c)
}

// GetLogsByCid returns history logs of a Cid.
func (s *Scheduler) GetLogsByCid(ctx context.Context, iid ffs.APIID, c cid.Cid) ([]ffs.LogEntry, error) {
	lgs, err := s.l.GetByCid(ctx, iid, c)
	if err != nil {
		return nil, fmt.Errorf("getting logs: %s", err)
	}
	return lgs, nil
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

// ListStorageJobsConfig controls the behavior for listing StorageJobs.
type ListStorageJobsConfig struct {
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

// ListStorageJobs lists StorageJobs according to the provided ListStorageJobsConfig.
func (s *Scheduler) ListStorageJobs(config ListStorageJobsConfig) ([]ffs.StorageJob, bool, string, error) {
	c := sjstore.ListConfig{
		APIIDFilter:   config.APIIDFilter,
		CidFilter:     config.CidFilter,
		Limit:         config.Limit,
		Ascending:     config.Ascending,
		Select:        sjstore.Select(config.Select),
		NextPageToken: config.NextPageToken,
	}
	return s.sjs.List(c)
}

// StorageConfig returns the storage config for a job.
func (s *Scheduler) StorageConfig(jid ffs.JobID) (ffs.StorageConfig, error) {
	a, err := s.as.GetStorageAction(jid)
	if err == astore.ErrNotFound {
		return ffs.StorageConfig{}, ErrNotFound
	}
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("getting storage action: %v", err)
	}
	return a.Cfg, nil
}

// executeStorage executes a Job. If an error is returned, it means that the Job
// should be considered failed. If error is nil, it still can return []ffs.DealError
// since some deals failing isn't necessarily a fatal Job config execution.
func (s *Scheduler) executeStorage(ctx context.Context, a astore.StorageAction, job ffs.StorageJob, dealUpdates chan deals.StorageDealInfo) (ffs.StorageInfo, []ffs.DealError, error) {
	ci, err := s.getRefreshedInfo(ctx, a.APIID, a.Cid)
	if err != nil {
		return ffs.StorageInfo{}, nil, fmt.Errorf("getting current cid info from store: %s", err)
	}

	if a.ReplacedCid.Defined() {
		if err := s.Untrack(a.APIID, a.ReplacedCid); err != nil && err != astore.ErrNotFound {
			return ffs.StorageInfo{}, nil, fmt.Errorf("untracking replaced cid: %s", err)
		}
	}

	var hot ffs.HotInfo
	if a.Cfg.Hot.Enabled {
		s.l.Log(ctx, "Executing Hot-Storage configuration...")
		hot, err = s.executeEnabledHotStorage(ctx, a.APIID, ci, a.Cfg.Hot, a.Cfg.Cold.Filecoin.Addr, a.ReplacedCid)
		if err != nil {
			s.l.Log(ctx, "Enabled Hot-Storage excution failed.")
			return ffs.StorageInfo{}, nil, fmt.Errorf("executing enabled hot-storage: %s", err)
		}
		s.l.Log(ctx, "Hot-Storage configuration ran successfully.")
	}

	// We want to avoid relying on Lotus working in online-mode.
	// We need to take care ourselves of pulling the data from
	// the IPFS network.
	if !a.Cfg.Hot.Enabled && a.Cfg.Cold.Enabled {
		s.l.Log(ctx, "Automatically staging Cid from the IPFS network...")
		stageCtx, cancel := context.WithTimeout(ctx, time.Duration(a.Cfg.Hot.Ipfs.AddTimeout)*time.Second)
		defer cancel()
		if err := s.hs.StageCid(stageCtx, a.APIID, a.Cid); err != nil {
			return ffs.StorageInfo{}, nil, fmt.Errorf("automatically staging cid: %s", err)
		}
	}

	s.l.Log(ctx, "Executing Cold-Storage configuration...")
	cold, errors, err := s.executeColdStorage(ctx, ci, a.Cfg, dealUpdates)
	if err != nil {
		s.l.Log(ctx, "Cold-Storage execution failed.")
		return ffs.StorageInfo{}, errors, fmt.Errorf("executing cold-storage config: %s", err)
	}
	s.l.Log(ctx, "Cold-Storage configuration ran successfully.")

	if !a.Cfg.Hot.Enabled {
		s.l.Log(ctx, "Executing Hot-Storage configuration...")
		if err := s.executeDisabledHotStorage(ctx, a.APIID, a.Cid); err != nil {
			s.l.Log(ctx, "Disabled Hot-Storage execution failed.")
			return ffs.StorageInfo{}, nil, fmt.Errorf("executing disabled hot-storage: %s", err)
		}
		s.l.Log(ctx, "Hot-Storage configuration ran successfully.")
	}

	return ffs.StorageInfo{
		APIID:   a.APIID,
		JobID:   job.ID,
		Cid:     a.Cid,
		Hot:     hot,
		Cold:    cold,
		Created: time.Now(),
	}, errors, nil
}

// ensureCorrectPinning ensures that the Cid has the correct pinning flag in hot storage.
func (s *Scheduler) executeDisabledHotStorage(ctx context.Context, iid ffs.APIID, c cid.Cid) error {
	ok, err := s.hs.IsPinned(ctx, iid, c)
	if err != nil {
		return fmt.Errorf("getting pinned status: %s", err)
	}
	if !ok {
		s.l.Log(ctx, "Data was already unpinned.")
		return nil
	}
	if err := s.hs.Unpin(ctx, iid, c); err != nil {
		return fmt.Errorf("unpinning cid %s: %s", c, err)
	}
	s.l.Log(ctx, "Data was unpinned.")

	return nil
}

// executeEnabledHotStorageEnabled runs the logic if the Job has Hot Storage enabled.
func (s *Scheduler) executeEnabledHotStorage(ctx context.Context, iid ffs.APIID, curr ffs.StorageInfo, cfg ffs.HotConfig, waddr string, replaceCid cid.Cid) (ffs.HotInfo, error) {
	if curr.Hot.Enabled {
		s.l.Log(ctx, "No actions needed in enabling Hot Storage.")
		return curr.Hot, nil
	}

	ipfsTimeout := time.Duration(cfg.Ipfs.AddTimeout) * time.Second
	sctx, cancel := context.WithTimeout(ctx, ipfsTimeout)
	defer cancel()

	var size int
	var err error
	if !replaceCid.Defined() {
		s.l.Log(ctx, "Fetching from the IPFS network...")
		size, err = s.hs.Pin(sctx, iid, curr.Cid)
	} else {
		s.l.Log(ctx, "Replace of previous pin %s", replaceCid)
		size, err = s.hs.Replace(sctx, iid, replaceCid, curr.Cid)
	}
	if err != nil {
		s.l.Log(ctx, "Direct fetching from IPFS wasn't possible.")
		if !cfg.AllowUnfreeze || len(curr.Cold.Filecoin.Proposals) == 0 {
			s.l.Log(ctx, "Unfreeze is disabled or active Filecoin deals are unavailable.")
			return ffs.HotInfo{}, fmt.Errorf("pinning cid in hot storage: %s", err)
		}
		s.l.Log(ctx, "Unfreezing from Filecoin...")
		if len(curr.Cold.Filecoin.Proposals) == 0 {
			return ffs.HotInfo{}, fmt.Errorf("no active deals to make retrieval possible")
		}
		var pieceCid *cid.Cid
		var miners []string
		for _, p := range curr.Cold.Filecoin.Proposals {
			if p.PieceCid != cid.Undef {
				pieceCid = &p.PieceCid
			}
			miners = append(miners, p.Miner)
		}

		fi, err := s.cs.Fetch(ctx, curr.Cold.Filecoin.DataCid, pieceCid, waddr, miners, cfg.UnfreezeMaxPrice, "")
		if err != nil {
			return ffs.HotInfo{}, fmt.Errorf("unfreezing from Cold Storage: %s", err)
		}
		s.l.Log(ctx, "Unfrozen successfully from %s with cost %d attoFil, saving in Hot-Storage...", fi.RetrievedMiner, fi.FundsSpent)
		size, err = s.hs.Pin(ctx, iid, curr.Cold.Filecoin.DataCid)
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

func (s *Scheduler) getRefreshedInfo(ctx context.Context, iid ffs.APIID, c cid.Cid) (ffs.StorageInfo, error) {
	var err error
	ci, err := s.cis.Get(iid, c)
	if err != nil {
		if err != cistore.ErrNotFound {
			return ffs.StorageInfo{}, ErrNotFound
		}
		return ffs.StorageInfo{Cid: c, APIID: iid}, nil // Default value has both storages disabled
	}

	ci.Hot, err = s.getRefreshedHotInfo(ctx, iid, c, ci.Hot)
	if err != nil {
		return ffs.StorageInfo{}, fmt.Errorf("getting refreshed hot info: %s", err)
	}

	ci.Cold, err = s.getRefreshedColdInfo(ctx, ci.Cold)
	if err != nil {
		return ffs.StorageInfo{}, fmt.Errorf("getting refreshed cold info: %s", err)
	}

	return ci, nil
}

func (s *Scheduler) getRefreshedHotInfo(ctx context.Context, iid ffs.APIID, c cid.Cid, curr ffs.HotInfo) (ffs.HotInfo, error) {
	var err error
	curr.Enabled, err = s.hs.IsPinned(ctx, iid, c)
	if err != nil {
		return ffs.HotInfo{}, err
	}
	return curr, nil
}

func (s *Scheduler) getRefreshedColdInfo(ctx context.Context, curr ffs.ColdInfo) (ffs.ColdInfo, error) {
	activeDeals := make([]ffs.FilStorage, 0, len(curr.Filecoin.Proposals))
	for _, fp := range curr.Filecoin.Proposals {
		_, err := s.cs.GetDealInfo(ctx, fp.DealID)
		if err == ffs.ErrOnChainDealNotFound {
			s.l.Log(ctx, "Detected that deal %d isn't active anymore, removing from active deals.", fp.DealID)
			continue
		}
		if err != nil {
			return ffs.ColdInfo{}, fmt.Errorf("getting deal %d state: %s", fp.DealID, err)
		}
		activeDeals = append(activeDeals, fp)
	}
	curr.Filecoin.Proposals = activeDeals
	return curr, nil
}

func (s *Scheduler) executeColdStorage(ctx context.Context, curr ffs.StorageInfo, cfg ffs.StorageConfig, dealUpdates chan deals.StorageDealInfo) (ffs.ColdInfo, []ffs.DealError, error) {
	if !cfg.Cold.Enabled {
		s.l.Log(ctx, "Cold-Storage was disabled, Filecoin deals will eventually expire.")
		return curr.Cold, nil, nil
	}
	curr.Cold.Enabled = true

	// 1. If we recognize there were some unfinished started deals, then
	// Powergate was closed while that was being executed. If that's the case
	// we resume tracking those deals until they finish.
	sds, err := s.sjs.GetStartedDeals(curr.APIID, curr.Cid)
	if err != nil {
		return ffs.ColdInfo{}, nil, fmt.Errorf("checking for started deals: %s", err)
	}
	var allErrors []ffs.DealError
	if len(sds) > 0 {
		s.l.Log(ctx, "Resuming %d dettached executing deals...", len(sds))
		okResumedDeals, failedResumedDeals := s.waitForDeals(ctx, curr.Cid, sds, dealUpdates)
		s.l.Log(ctx, "A total of %d resumed deals finished successfully", len(okResumedDeals))
		allErrors = append(allErrors, failedResumedDeals...)
		// Append the resumed and confirmed deals to the current active proposals
		curr.Cold.Filecoin.Proposals = append(okResumedDeals, curr.Cold.Filecoin.Proposals...)

		// We can already clean resumed started deals.
		if err := s.sjs.RemoveStartedDeals(curr.APIID, curr.Cid); err != nil {
			return ffs.ColdInfo{}, allErrors, fmt.Errorf("removing resumed started deals storage: %s", err)
		}
	}

	currentEpoch, err := s.cs.GetCurrentEpoch(ctx)
	if err != nil {
		log.Error(err)
	} else {
		for _, deal := range curr.Cold.Filecoin.Proposals {
			// no need to alert for renewed deals
			if deal.Renewed {
				continue
			}

			s.notifier.Alert(
				notifications.DealExpirationAlert{
					JobID:        curr.JobID,
					DealID:       deal.DealID,
					PieceCid:     deal.PieceCid,
					Miner:        deal.Miner,
					ExpiryEpoch:  deal.StartEpoch + uint64(deal.Duration),
					CurrentEpoch: currentEpoch,
				},
				cfg.Notifications,
			)
		}
	}

	// 2. If this Storage Config is renewable, then let's check if any of the existing deals
	// should be renewed, and do it.
	if cfg.Cold.Filecoin.Renew.Enabled {
		if curr.Hot.Enabled {
			s.l.Log(ctx, "Checking deal renewals...")
			newFilInfo, errors, err := s.cs.EnsureRenewals(ctx, curr.Cid, curr.Cold.Filecoin, cfg.Cold.Filecoin, s.dealFinalityTimeout, dealUpdates)
			if err != nil {
				s.l.Log(ctx, "Deal renewal process couldn't be executed: %s", err)
			} else {
				for _, e := range errors {
					s.l.Log(ctx, "Deal deal renewal errored. ProposalCid: %s, Miner: %s, Cause: %s", e.ProposalCid, e.Miner, e.Message)
				}
				numDeals := len(newFilInfo.Proposals) - len(curr.Cold.Filecoin.Proposals)
				if numDeals > 0 {
					// If renew process created deals, we eagerly save this information in the datastore.
					// Further work about the new storage config could decide the Job failed and we'd lose
					// this information if not saved.
					if err := s.cis.Put(curr); err != nil {
						return ffs.ColdInfo{}, nil, fmt.Errorf("eager saving of new info: %s", err)
					}
					s.l.Log(ctx, "A total of %d new deals were created in the renewal process", numDeals)
				}
				s.l.Log(ctx, "Deal renewal evaluated successfully")
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
	if s.sr2RepFactor != nil {
		cfg.Cold.Filecoin.RepFactor, err = s.sr2RepFactor()
		if err != nil {
			return ffs.ColdInfo{}, nil, fmt.Errorf("getting SR2 replication factor: %s", err)
		}
	}
	if cfg.Cold.Filecoin.RepFactor-len(curr.Cold.Filecoin.Proposals) <= 0 {
		s.l.Log(ctx, "The current replication factor is equal or higher than desired, avoiding making new deals.")
		return curr.Cold, nil, nil
	}

	// The answer is yes, calculate how many extra deals we need and create them.
	deltaFilConfig := createDeltaFilConfig(cfg.Cold, curr.Cold.Filecoin)
	s.l.Log(ctx, "Current replication factor is lower than desired, making %d new deals...", deltaFilConfig.RepFactor)
	startedProposals, rejectedProposals, size, err := s.cs.Store(ctx, curr.Cid, deltaFilConfig)
	if err != nil {
		s.l.Log(ctx, "Starting deals failed, with cause: %s", err)
		return ffs.ColdInfo{}, rejectedProposals, err
	}
	allErrors = append(allErrors, rejectedProposals...)

	// If *none* of the tried proposals succeeded, then the Job fails.
	if len(startedProposals) == 0 {
		return ffs.ColdInfo{}, allErrors, fmt.Errorf("all proposals were rejected")
	}

	// Track all deals that weren't rejected, just in case Powergate crashes/closes before
	// we see them finalize, so they can be detected and resumed on starting Powergate again (point 1. above)
	if err := s.sjs.AddStartedDeals(curr.APIID, curr.Cid, startedProposals); err != nil {
		return ffs.ColdInfo{}, rejectedProposals, err
	}

	// Wait for started deals.
	okDeals, failedDeals := s.waitForDeals(ctx, curr.Cid, startedProposals, dealUpdates)
	allErrors = append(allErrors, failedDeals...)
	if err := s.sjs.RemoveStartedDeals(curr.APIID, curr.Cid); err != nil {
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
			Size:      uint64(size),
			Proposals: append(okDeals, curr.Cold.Filecoin.Proposals...), // Append to any existing other proposals
		},
	}, allErrors, nil
}

func (s *Scheduler) waitForDeals(ctx context.Context, c cid.Cid, startedProposals []cid.Cid, dealUpdates chan deals.StorageDealInfo) ([]ffs.FilStorage, []ffs.DealError) {
	s.l.Log(ctx, "Watching deals unfold...")

	var failedDeals []ffs.DealError
	var okDeals []ffs.FilStorage
	var wg sync.WaitGroup
	var lock sync.Mutex
	wg.Add(len(startedProposals))
	for _, pc := range startedProposals {
		pc := pc
		go func() {
			defer wg.Done()

			res, err := s.cs.WaitForDeal(ctx, c, pc, s.dealFinalityTimeout, dealUpdates)
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
