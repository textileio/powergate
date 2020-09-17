package api

import (
	"context"
	"fmt"

	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

// GetStorageJob returns the current state of the specified job.
func (i *API) GetStorageJob(ctx context.Context, jid ffs.JobID) (ffs.StorageJob, error) {
	job, err := i.sched.GetJob(jid)
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
		j, err := i.sched.GetJob(jid)
		if err == scheduler.ErrNotFound {
			return fmt.Errorf("job id %s not found", jid)
		}
		if err != nil {
			return fmt.Errorf("getting current job state: %s", err)
		}
		jobs = append(jobs, j)
	}

	ch := make(chan ffs.StorageJob, 1)
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
