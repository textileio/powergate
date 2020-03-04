package scheduler

import (
	"github.com/textileio/fil-tools/fpa"
)

// JobStore persist Job information and allows to watch for Job state updates.
type JobStore interface {
	GetByStatus(fpa.JobStatus) ([]fpa.Job, error)
	Put(fpa.Job) error

	Watch(fpa.InstanceID) <-chan fpa.Job
	Unwatch(<-chan fpa.Job)
}
