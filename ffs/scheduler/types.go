package scheduler

import (
	"context"
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
	Queue(ffs.Job) error
	// Dequeue returns the next queued Job. It automatically changes
	// its state to in-progress.
	Dequeue() (*ffs.Job, error)

	// ChangeStatus changes the status of a Job.
	ChangeStatus(ffs.JobID, ffs.JobStatus) error
	// Get retrieves job data from the store.
	Get(ffs.JobID) (ffs.Job, error)
	// Watch subscribes to all job state changes within an instance.
	Watch(context.Context, chan<- ffs.Job, ffs.APIID) error
}

// Action represents an action to be executed by the Scheduler.
type Action struct {
	APIID       ffs.APIID
	Waddr       string
	Cfg         ffs.CidConfig
	ReplacedCid cid.Cid
}

// ActionStore persist actions for Cids.
type ActionStore interface {
	// Put saves a new state for a Job.
	Put(ffs.JobID, Action) error
	// Get returns the current state of a Job.
	Get(ffs.JobID) (Action, error)
	// Remove removes the action associated with a Cid.
	Remove(cid.Cid) error
	// GetRenewable returns the known pushed configs that have enabled
	// renew Filecoin flag for their deals.
	GetRenewable() ([]Action, error)
}

// CidInfoStore persists CidInfo which represent the current storage
// state of a Cid.
type CidInfoStore interface {
	Put(ffs.CidInfo) error
	Get(cid.Cid) (ffs.CidInfo, error)
}
