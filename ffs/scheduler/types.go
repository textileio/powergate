package scheduler

import (
	"errors"

	"github.com/textileio/powergate/ffs"
)

var (
	ErrNotFound = errors.New("job not found")
)

// JobStore persist Job information and allows to watch for Job state updates.
type JobStore interface {
	// GetByStatus returns jobs with a particular status.
	GetByStatus(ffs.JobStatus) ([]ffs.Job, error)
	// Put saves a new state for a Job.
	Put(ffs.Job) error
	// Get returns the current state of a Job.
	Get(ffs.JobID) (ffs.Job, error)

	// Watch subscribes to all job state changes within an instance.
	Watch(ffs.InstanceID) <-chan ffs.Job
	// Unwatch unregisters a Watch subscriber.
	Unwatch(<-chan ffs.Job)
}
