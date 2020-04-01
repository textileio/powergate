package scheduler

import (
	"context"
	"fmt"
	"io"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("ffs-scheduler")
)

// Scheduler receives actions to store a Cid in Hot and Cold layers. These actions are
// created as Jobs which have a lifecycle that can be watched by external actors.
// This Jobs are executed by delegating the work to the Hot and Cold layers configured for
// the scheduler.
type Scheduler struct {
	cs  ffs.ColdStorage
	hs  ffs.HotStorage
	js  JobStore
	pcs PushConfigStore
	cis CidInfoStore

	queuedWork chan struct{}

	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

var _ ffs.Scheduler = (*Scheduler)(nil)

// New returns a new instance of Scheduler which uses JobStore as its backing repository for state,
// HotStorage for the hot layer, and ColdStorage for the cold layer.
func New(js JobStore, pcs PushConfigStore, cis CidInfoStore, hs ffs.HotStorage, cs ffs.ColdStorage) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	sch := &Scheduler{
		cs:  cs,
		hs:  hs,
		js:  js,
		pcs: pcs,
		cis: cis,

		queuedWork: make(chan struct{}, 1),

		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
	go sch.run()
	return sch
}

// PushConfig queues the specified CidConfig to be executed as a new Job. It returns
// the created JobID for further tracking of its state.
func (s *Scheduler) PushConfig(action ffs.PushConfigAction) (ffs.JobID, error) {
	if !action.Config.Cid.Defined() {
		return ffs.EmptyJobID, fmt.Errorf("cid can't be undefined")
	}
	jid := ffs.NewJobID()
	j := ffs.Job{
		ID:         jid,
		InstanceID: action.InstanceID,
		Status:     ffs.Queued,
	}
	if err := s.js.Put(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving push config action in store: %s", err)
	}
	if err := s.pcs.Put(j.ID, action); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving pushed config in store: %s", err)
	}
	select {
	case s.queuedWork <- struct{}{}:
	default:
	}
	return jid, nil
}

// GetCidInfo returns the current storage state of a Cid. Returns ErrNotFound
// if there isn't information for a Cid.
func (s *Scheduler) GetCidInfo(c cid.Cid) (ffs.CidInfo, error) {
	info, err := s.cis.Get(c)
	if err == ErrNotFound {
		return ffs.CidInfo{}, err
	}
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting CidInfo from store: %s", err)
	}
	return info, nil
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
		if err == ErrNotFound {
			return ffs.Job{}, err
		}
		return ffs.Job{}, fmt.Errorf("get Job from store: %s", err)
	}
	return j, nil
}

// Watch returns a channel to listen to Job status changes from a specified
// Api instance. It immediately pushes the current Job state to the channel.
func (s *Scheduler) Watch(iid ffs.ApiID) <-chan ffs.Job {
	return s.js.Watch(iid)
}

// Unwatch unregisters a subscribing channel created by Watch().
func (s *Scheduler) Unwatch(ch <-chan ffs.Job) {
	s.js.Unwatch(ch)
}

// Close terminates the scheduler.
func (s *Scheduler) Close() error {
	s.cancel()
	<-s.finished
	return nil
}

func (s *Scheduler) run() {
	defer close(s.finished)
	for {
		select {
		case <-s.ctx.Done():
			log.Infof("terminating scheduler daemon")
			return
		case <-time.After(util.AvgBlockTime):
			s.scanRenewable(s.ctx)
		case <-s.queuedWork:
			s.executeQueuedJobs(s.ctx)
		}
	}
}

func (s *Scheduler) scanRenewable(ctx context.Context) {
	renewableActions, err := s.pcs.GetRenewable()
	if err != nil {
		log.Errorf("getting renweable cid configs from store: %s", err)
	}
	for _, a := range renewableActions {
		if err := s.evaluateRenewal(ctx, a); err != nil {
			log.Errorf("renweal of %s: %s", a.Config.Cid, err)
		}
	}
}

func (s *Scheduler) evaluateRenewal(ctx context.Context, a ffs.PushConfigAction) error {
	inf, err := s.cis.Get(a.Config.Cid)
	if err == ErrNotFound {
		log.Infof("skip renewal evaluation for %s since Cid isn't stored yet", a.Config.Cid)
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting cid info from store: %s", err)
	}

	inf.Cold.Filecoin, err = s.cs.EnsureRenewals(ctx, a.Config.Cid, inf.Cold.Filecoin, a.WalletAddr, a.Config.Cold.Filecoin)
	if err != nil {
		return fmt.Errorf("evaluating renewal in cold-storage: %s", err)
	}

	if err := s.cis.Put(inf); err != nil {
		return fmt.Errorf("saving new cid info in store: %s", err)
	}

	return nil
}

func (s *Scheduler) executeQueuedJobs(ctx context.Context) {
	js, err := s.js.GetByStatus(ffs.Queued)
	if err != nil {
		log.Errorf("getting queued jobs: %s", err)
		return
	}
	log.Infof("detected %d queued jobs", len(js))
	for _, j := range js {
		log.Infof("executing job %s", j.ID)
		if err := s.mutateJobStatus(j, ffs.InProgress); err != nil {
			log.Errorf("changing job to in-progress: %s", err)
			return
		}
		info, err := s.execute(s.ctx, j)
		if err != nil {
			log.Errorf("executing job %s: %s", j.ID, err)
			j.ErrCause = err.Error()
			if err := s.mutateJobStatus(j, ffs.Failed); err != nil {
				log.Errorf("changing job to failed: %s", err)
			}
			return
		}
		if err := s.cis.Put(info); err != nil {
			log.Errorf("saving cid info to store: %s", err)
			return
		}
		if err := s.mutateJobStatus(j, ffs.Success); err != nil {
			log.Errorf("changing job to success: %s", err)
			return
		}
		log.Infof("job %s executed with final state %d and errcause %q", j.ID, j.Status, j.ErrCause)

		if ctx.Err() != nil {
			break
		}
	}
}

func (s *Scheduler) execute(ctx context.Context, job ffs.Job) (ffs.CidInfo, error) {
	a, err := s.pcs.Get(job.ID)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting push config action data from store: %s", err)
	}

	ci, err := s.getRefreshedInfo(ctx, a.Config.Cid)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting current cid info from store: %s", err)
	}

	hot, err := s.executeHotStorage(ctx, ci, a.Config.Hot, a.WalletAddr)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("executing hot-storage config: %s", err)
	}

	cold, err := s.executeColdStorage(ctx, ci, a.Config.Cold, a.WalletAddr)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("executing cold-storage config: %s", err)
	}

	return ffs.CidInfo{
		JobID:   job.ID,
		Cid:     a.Config.Cid,
		Hot:     hot,
		Cold:    cold,
		Created: time.Now(),
	}, nil
}

func (s *Scheduler) executeHotStorage(ctx context.Context, curr ffs.CidInfo, cfg ffs.HotConfig, waddr string) (ffs.HotInfo, error) {
	if cfg.Enabled == curr.Hot.Enabled {
		return curr.Hot, nil
	}

	if !cfg.Enabled {
		if err := s.hs.Remove(ctx, curr.Cid); err != nil {
			return ffs.HotInfo{}, fmt.Errorf("removing from hot storage: %s", err)
		}
		return ffs.HotInfo{Enabled: false}, nil
	}

	hotPinCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.Ipfs.AddTimeout))
	defer cancel()
	size, err := s.hs.Pin(hotPinCtx, curr.Cid)
	if err != nil {
		if !cfg.AllowUnfreeze || len(curr.Cold.Filecoin.Proposals) == 0 {
			return ffs.HotInfo{}, fmt.Errorf("pinning cid in hot storage: %s", err)
		}
		bs := &hotStorageBlockstore{ctx: ctx, put: s.hs.Put}
		carHeaderCid, err := s.cs.Retrieve(ctx, curr.Cold.Filecoin.DataCid, bs, waddr)
		if err != nil {
			return ffs.HotInfo{}, fmt.Errorf("unfreezing from Cold Storage: %s", err)
		}
		size, err = s.hs.Pin(ctx, carHeaderCid)
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
		if err != ErrNotFound {
			return ffs.CidInfo{}, fmt.Errorf("getting current cid info from store: %s", err)
		}
		return ffs.CidInfo{Cid: c}, nil // Default value has both storages disabled
	}

	ci.Hot, err = s.getRefreshedHotInfo(ctx, c, ci.Hot)
	if err != nil {
		return ffs.CidInfo{}, fmt.Errorf("getting refreshed hot info: %s", err)
	}

	ci.Cold, err = s.getRefreshedColdInfo(ctx, c, ci.Cold)
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

func (s *Scheduler) getRefreshedColdInfo(ctx context.Context, c cid.Cid, curr ffs.ColdInfo) (ffs.ColdInfo, error) {
	activeDeals := make([]ffs.FilStorage, 0, len(curr.Filecoin.Proposals))
	for _, fp := range curr.Filecoin.Proposals {
		active, err := s.cs.IsFilDealActive(ctx, fp.ProposalCid)
		if err != nil {
			return ffs.ColdInfo{}, fmt.Errorf("getting deal state of proposal %s: %s", fp.ProposalCid, err)
		}
		if active {
			activeDeals = append(activeDeals, fp)
		}
	}
	curr.Filecoin.Proposals = activeDeals
	return curr, nil
}

func (s *Scheduler) executeColdStorage(ctx context.Context, curr ffs.CidInfo, cfg ffs.ColdConfig, waddr string) (ffs.ColdInfo, error) {
	if !cfg.Enabled {
		return curr.Cold, nil
	}

	if isCurrentRepFactorEnough(cfg.Filecoin.RepFactor, curr) {
		log.Infof("replication well enough, avoid making new deals")
		return curr.Cold, nil

	}

	deltaFilConfig := createDeltaFilConfig(cfg, curr.Cold.Filecoin)
	fi, err := s.cs.Store(ctx, curr.Cid, waddr, deltaFilConfig)
	if err != nil {
		return ffs.ColdInfo{}, err
	}
	return ffs.ColdInfo{
		Filecoin: fi,
	}, nil
}

func isCurrentRepFactorEnough(desiredRepFactor int, curr ffs.CidInfo) bool {
	return desiredRepFactor-len(curr.Cold.Filecoin.Proposals) <= 0

}

func createDeltaFilConfig(cfg ffs.ColdConfig, curr ffs.FilInfo) ffs.FilConfig {
	res := cfg.Filecoin
	res.RepFactor = cfg.Filecoin.RepFactor - len(curr.Proposals)
	for _, p := range curr.Proposals {
		res.ExcludedMiners = append(res.ExcludedMiners, p.Miner)
	}
	return res
}

func (s *Scheduler) mutateJobStatus(j ffs.Job, status ffs.JobStatus) error {
	j.Status = status
	if err := s.js.Put(j); err != nil {
		return err
	}
	return nil
}

type hotStorageBlockstore struct {
	ctx context.Context
	put func(context.Context, blocks.Block) error
}

func (hsb *hotStorageBlockstore) Put(b blocks.Block) error {
	if err := hsb.put(hsb.ctx, b); err != nil {
		return fmt.Errorf("saving block in hot-storage: %s", err)
	}
	return nil
}
