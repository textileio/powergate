package astore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("ffs-sched-astore")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("action not found")
)

// StorageAction contains information necessary to a Job execution.
type StorageAction struct {
	APIID       ffs.APIID
	Cid         cid.Cid
	Cfg         ffs.StorageConfig
	ReplacedCid cid.Cid
}

// Store persists Actions.
type Store struct {
	ds datastore.Datastore
}

// New returns a new ActionStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Get gets an action for a JobID. If doesn't exist, returns ErrNotFound.
func (s *Store) Get(jid ffs.JobID) (StorageAction, error) {
	var a StorageAction
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return a, ErrNotFound
	}
	if err != nil {
		return a, fmt.Errorf("get from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &a); err != nil {
		return a, fmt.Errorf("unmarshaling from datastore: %s", err)
	}
	return a, nil
}

// Put saves a new Action for a Job.
func (s *Store) Put(jid ffs.JobID, a StorageAction) error {
	buf, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("json marshaling: %s", err)
	}
	if err := s.ds.Put(makeKey(jid), buf); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

func makeKey(jid ffs.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
