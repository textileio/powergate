package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/astore"
	"github.com/textileio/powergate/ffs/scheduler/internal/cistore"
	"github.com/textileio/powergate/ffs/scheduler/internal/sjstore"
)

var (
	// HardcodedHotTimeout is a temporary override of storage configs
	// value for AddTimeout.
	HardcodedHotTimeout = time.Second * 300
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
	if jid := s.sjs.GetExecutingJob(c); jid != nil {
		s.l.Log(ctx, "Job %s is already being executed for the same data, this job will be queued until it finishes or is canceled.")
	}

	select {
	case s.sd.evaluateQueue <- struct{}{}:
	default:
	}

	s.l.Log(ctx, "Configuration saved successfully")
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

// GetJob the current state of a Job.
func (s *Scheduler) GetJob(jid ffs.JobID) (ffs.StorageJob, error) {
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
func (s *Scheduler) GetLogsByCid(ctx context.Context, c cid.Cid) ([]ffs.LogEntry, error) {
	lgs, err := s.l.GetByCid(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting logs: %s", err)
	}
	return lgs, nil
}

// executeStorage executes a Job. If an error is returned, it means that the Job
// should be considered failed. If error is nil, it still can return []ffs.DealError
// since some deals failing isn't necessarily a fatal Job config execution.
func (s *Scheduler) executeStorage(ctx context.Context, a astore.StorageAction, job ffs.StorageJob) (ffs.CidInfo, []ffs.DealError, error) {
	ci, err := s.getRefreshedInfo(ctx, a.Cid)
	if err != nil {
		return ffs.CidInfo{}, nil, fmt.Errorf("getting current cid info from store: %s", err)
	}

	if a.ReplacedCid.Defined() {
		if err := s.Untrack(a.ReplacedCid); err != nil && err != astore.ErrNotFound {
			return ffs.CidInfo{}, nil, fmt.Errorf("untracking replaced cid: %s", err)
		}
	}

	s.l.Log(ctx, "Ensuring Hot-Storage satisfies the configuration...")
	hot, err := s.executeHotStorage(ctx, ci, a.Cfg.Hot, a.Cfg.Cold.Filecoin.Addr, a.ReplacedCid)
	if err != nil {
		s.l.Log(ctx, "Hot-Storage excution failed.")
		return ffs.CidInfo{}, nil, fmt.Errorf("executing hot-storage config: %s", err)
	}
	s.l.Log(ctx, "Hot-Storage execution ran successfully.")

	s.l.Log(ctx, "Ensuring Cold-Storage satisfies the configuration...")
	cold, errors, err := s.executeColdStorage(ctx, ci, a.Cfg.Cold)
	if err != nil {
		s.l.Log(ctx, "Cold-Storage execution failed.")
		return ffs.CidInfo{}, errors, fmt.Errorf("executing cold-storage config: %s", err)
	}
	s.l.Log(ctx, "Cold-Storage execution ran successfully.")

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
		s.l.Log(ctx, "No actions needed in Hot Storage.")
		return curr.Hot, nil
	}

	if !cfg.Enabled {
		if err := s.hs.Remove(ctx, curr.Cid); err != nil {
			return ffs.HotInfo{}, fmt.Errorf("removing from hot storage: %s", err)
		}
		s.l.Log(ctx, "Cid successfully removed from Hot Storage.")
		return ffs.HotInfo{Enabled: false}, nil
	}

	// ToDo: this is a hot-fix to force a big timeout until we have a
	// migration tool to make this tunable again.
	sctx, cancel := context.WithTimeout(ctx, HardcodedHotTimeout)
	defer cancel()

	var size int
	var err error
	if !replaceCid.Defined() {
		size, err = s.hs.Store(sctx, curr.Cid)
	} else {
		s.l.Log(ctx, "Replace of previous pin %s", replaceCid)
		size, err = s.hs.Replace(sctx, replaceCid, curr.Cid)
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
		s.l.Log(ctx, "Cold-Storage was disabled, Filecoin deals will eventually expire.")
		return curr.Cold, nil, nil
	}

	// 1. If we recognize there were some unfinished started deals, then
	// Powergate was closed while that was being executed. If that's the case
	// we resume tracking those deals until they finish.
	sds, err := s.sjs.GetStartedDeals(curr.Cid)
	if err != nil {
		return ffs.ColdInfo{}, nil, fmt.Errorf("checking for started deals: %s", err)
	}
	var allErrors []ffs.DealError
	if len(sds) > 0 {
		s.l.Log(ctx, "Resuming %d dettached executing deals...", len(sds))
		okResumedDeals, failedResumedDeals := s.waitForDeals(ctx, curr.Cid, sds)
		s.l.Log(ctx, "A total of %d resumed deals finished successfully", len(okResumedDeals))
		allErrors = append(allErrors, failedResumedDeals...)
		// Append the resumed and confirmed deals to the current active proposals
		curr.Cold.Filecoin.Proposals = append(okResumedDeals, curr.Cold.Filecoin.Proposals...)
	}

	// 2. If this Storage Config is renewable, then let's check if any of the existing deals
	// should be renewed, and do it.
	if cfg.Filecoin.Renew.Enabled {
		if curr.Hot.Enabled {
			s.l.Log(ctx, "Checking deal renewals...")
			newFilInfo, errors, err := s.cs.EnsureRenewals(ctx, curr.Cid, curr.Cold.Filecoin, cfg.Filecoin, s.dealFinalityTimeout)
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
		cfg.Filecoin.RepFactor, err = s.sr2RepFactor()
		if err != nil {
			return ffs.ColdInfo{}, nil, fmt.Errorf("getting SR2 replication factor: %s", err)
		}
	}
	if cfg.Filecoin.RepFactor-len(curr.Cold.Filecoin.Proposals) <= 0 {
		s.l.Log(ctx, "The current replication factor is equal or higher than desired, avoiding making new deals.")
		return curr.Cold, nil, nil
	}

	// The answer is yes, calculate how many extra deals we need and create them.
	deltaFilConfig := createDeltaFilConfig(cfg, curr.Cold.Filecoin)
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
	if err := s.sjs.AddStartedDeals(curr.Cid, startedProposals); err != nil {
		return ffs.ColdInfo{}, rejectedProposals, err
	}

	// Wait for started deals.
	okDeals, failedDeals := s.waitForDeals(ctx, curr.Cid, startedProposals)
	allErrors = append(allErrors, failedDeals...)
	if err := s.sjs.RemoveStartedDeals(curr.Cid); err != nil {
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

			res, err := s.cs.WaitForDeal(ctx, c, pc, s.dealFinalityTimeout)
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
