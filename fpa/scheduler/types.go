package scheduler

import (
	"github.com/textileio/fil-tools/fpa"
)

type JobStore interface {
	GetByStatus(fpa.JobStatus) ([]fpa.Job, error)
	Put(fpa.Job) error

	Watch(fpa.InstanceID) <-chan fpa.Job
	Unwatch(<-chan fpa.Job)
}
