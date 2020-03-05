package scheduler

import (
	"errors"

	"github.com/textileio/fil-tools/fpa"
)

var (
	ErrNotFound = errors.New("job not found")
)

// JobStore persist Job information and allows to watch for Job state updates.
type JobStore interface {
	// GetByStatus returns jobs with a particular status.
	GetByStatus(fpa.JobStatus) ([]fpa.Job, error)
	// Put saves a new state for a Job.
	Put(fpa.Job) error
	// Get returns the current state of a Job.
	Get(fpa.JobID) (fpa.Job, error)

	// Watch subscribes to all job state changes within an instance.
	Watch(fpa.InstanceID) <-chan fpa.Job
	// Unwatch unregisters a Watch subscriber.
	Unwatch(<-chan fpa.Job)
}
