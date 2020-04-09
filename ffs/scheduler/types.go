package scheduler

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound is returned when an item isn't found on a Store.
	ErrNotFound = errors.New("item not found")
)

// JobStore persist Job information and allows to watch for Job state updates.
type JobStore interface {
	// Put saves job data in the store.
	Put(ffs.Job) error
	// Get retrieves job data from the store.
	Get(ffs.JobID) (ffs.Job, error)
	// GetByStatus returns jobs with a particular status.
	GetByStatus(ffs.JobStatus) ([]ffs.Job, error)

	// Watch subscribes to all job state changes within an instance.
	Watch(ffs.ApiID) <-chan ffs.Job
	// Unwatch unregisters a Watch subscriber.
	Unwatch(<-chan ffs.Job)
}

// PushConfigStore persist PushConfigActions for Cids.
type PushConfigStore interface {
	// Put saves a new state for a Job.
	Put(ffs.JobID, ffs.PushConfigAction) error
	// Get returns the current state of a Job.
	Get(ffs.JobID) (ffs.PushConfigAction, error)
	// GetRenewable returns the known pushed configs that have enabled
	// renew Filecoin flag for their deals.
	GetRenewable() ([]ffs.PushConfigAction, error)
}

// CidInfoStore persists CidInfo which represent the current storage
// state of a Cid.
type CidInfoStore interface {
	Put(ffs.CidInfo) error
	Get(cid.Cid) (ffs.CidInfo, error)
}
