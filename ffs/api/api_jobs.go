package api

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

// GetStorageJob returns the current state of the specified job.
func (i *API) GetStorageJob(ctx context.Context, jid ffs.JobID) (ffs.StorageJob, error) {
	job, err := i.sched.GetStorageJob(jid)
	if err == scheduler.ErrNotFound {
		return ffs.StorageJob{}, fmt.Errorf("job id %s not found", jid)
	}
	if err != nil {
		return ffs.StorageJob{}, fmt.Errorf("getting current job state: %s", err)
	}
	return job, nil
}

// WatchJobs subscribes to Job status changes. If jids is empty, it subscribes to
// all Job status changes corresonding to the instance. If jids is not empty,
// it immediately sends current state of those Jobs. If empty, it doesn't.
func (i *API) WatchJobs(ctx context.Context, c chan<- ffs.StorageJob, jids ...ffs.JobID) error {
	var jobs []ffs.StorageJob
	for _, jid := range jids {
		j, err := i.sched.GetStorageJob(jid)
		if err == scheduler.ErrNotFound {
			return fmt.Errorf("job id %s not found", jid)
		}
		if err != nil {
			return fmt.Errorf("getting current job state: %s", err)
		}
		jobs = append(jobs, j)
	}

	ch := make(chan ffs.StorageJob, 10)
	for _, j := range jobs {
		select {
		case ch <- j:
		default:
			log.Warnf("dropped notifying current job state on slow receiver on %s", i.cfg.ID)
		}
	}
	var err error
	go func() {
		err = i.sched.WatchJobs(ctx, ch, i.cfg.ID)
		close(ch)
	}()
	for j := range ch {
		if len(jids) == 0 {
			c <- j
		}
	JidLoop:
		for _, jid := range jids {
			if jid == j.ID {
				c <- j
				break JidLoop
			}
		}
	}
	if err != nil {
		return fmt.Errorf("scheduler listener: %s", err)
	}

	return nil
}

// GetQueuedStorageJobs returns queued jobs for the specified cids.
// If no cids are provided, data for all data cids is returned.
func (i *API) GetQueuedStorageJobs(cids ...cid.Cid) []ffs.StorageJob {
	return i.sched.GetQueuedStorageJobs(i.cfg.ID, cids...)
}

// GetExecutingStorageJobs returns executing jobs for the specified cids.
// If no cids are provided, data for all data cids is returned.
func (i *API) GetExecutingStorageJobs(cids ...cid.Cid) []ffs.StorageJob {
	return i.sched.GetExecutingStorageJobs(i.cfg.ID, cids...)
}

// GetLastFinalStorageJobs returns the most recent finished jobs for the specified cids.
// If no cids are provided, data for all data cids is returned.
func (i *API) GetLastFinalStorageJobs(cids ...cid.Cid) []ffs.StorageJob {
	return i.sched.GetLastFinalStorageJobs(i.cfg.ID, cids...)
}

// GetLastSuccessfulStorageJobs returns the most recent successful jobs for the specified cids.
// If no cids are provided, data for all data cids is returned.
func (i *API) GetLastSuccessfulStorageJobs(cids ...cid.Cid) []ffs.StorageJob {
	return i.sched.GetLastSuccessfulStorageJobs(i.cfg.ID, cids...)
}
