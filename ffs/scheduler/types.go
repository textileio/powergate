package scheduler

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

var (
	ErrNotFound = errors.New("job not found")
)

// Store persist Job information and allows to watch for Job state updates.
type JobStore interface {
	// PutJob saves job data in the store.
	Put(ffs.Job) error
	// GetJob retrieves job data from the store.
	Get(ffs.JobID) (ffs.Job, error)
	// GetByStatus returns jobs with a particular status.
	GetByStatus(ffs.JobStatus) ([]ffs.Job, error)

	// Watch subscribes to all job state changes within an instance.
	Watch(ffs.InstanceID) <-chan ffs.Job
	// Unwatch unregisters a Watch subscriber.
	Unwatch(<-chan ffs.Job)
}

type PushConfigStore interface {
	// PutPushConfigJob saves a new state for a Job.
	Put(ffs.JobID, ffs.PushConfigAction) error
	// Get returns the current state of a Job.
	Get(ffs.JobID) (ffs.PushConfigAction, error)
}

type CidInfoStore interface {
	Put(ffs.CidInfo) error
	Get(cid.Cid) (ffs.CidInfo, error)
}
