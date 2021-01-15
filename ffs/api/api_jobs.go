package api

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

// GetStorageJob returns the current state of the specified job.
func (i *API) GetStorageJob(jid ffs.JobID) (ffs.StorageJob, error) {
	job, err := i.sched.StorageJob(jid)
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
		j, err := i.sched.StorageJob(jid)
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

// ListStorageJobsSelect specifies which StorageJobs to list.
type ListStorageJobsSelect int

const (
	// All lists all StorageJobs and is the default.
	All ListStorageJobsSelect = iota
	// Queued lists queued StorageJobs.
	Queued
	// Executing lists executing StorageJobs.
	Executing
	// Final lists final StorageJobs.
	Final
)

// ListStorageJobsConfig controls the behavior for listing StorageJobs.
type ListStorageJobsConfig struct {
	// CidFilter filters StorageJobs list to the specified cid. Defaults to no filter.
	CidFilter cid.Cid
	// Limit limits the number of StorageJobs returned. Defaults to no limit.
	Limit uint64
	// Ascending returns the StorageJobs ascending by time. Defaults to false, descending.
	Ascending bool
	// Select specifies to return StorageJobs in the specified state.
	Select ListStorageJobsSelect
	// NextPageToken sets the slug from which to start building the next page of results.
	NextPageToken string
}

// ListStorageJobs lists StorageJobs according to the provided ListStorageJobsConfig.
func (i *API) ListStorageJobs(config ListStorageJobsConfig) ([]ffs.StorageJob, bool, string, error) {
	c := scheduler.ListStorageJobsConfig{
		APIIDFilter:   i.cfg.ID,
		CidFilter:     config.CidFilter,
		Limit:         config.Limit,
		Ascending:     config.Ascending,
		Select:        scheduler.Select(config.Select),
		NextPageToken: config.NextPageToken,
	}
	return i.sched.ListStorageJobs(c)
}

// StorageConfigForJob returns the StorageConfig associated with the specified job.
func (i *API) StorageConfigForJob(jid ffs.JobID) (ffs.StorageConfig, error) {
	sc, err := i.sched.StorageConfig(jid)
	if err == scheduler.ErrNotFound {
		return ffs.StorageConfig{}, ErrNotFound
	}
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("getting storage config for job: %v", err)
	}
	return sc, nil
}
