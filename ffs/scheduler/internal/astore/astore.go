package astore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/v2/ffs"
)

var (
	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("action not found")

	dsBaseStorageAction   = datastore.NewKey("storageaction")
	dsBaseRetrievalAction = datastore.NewKey("retrievalaction")
)

// StorageAction contains information necessary to execute a
// StorageJob.
type StorageAction struct {
	APIID       ffs.APIID
	Cid         cid.Cid
	Cfg         ffs.StorageConfig
	ReplacedCid cid.Cid
}

// RetrievalAction contains information necessary to execute a
// RetrievalJob.
type RetrievalAction struct {
	APIID         ffs.APIID
	RetrievalID   ffs.RetrievalID
	PayloadCid    cid.Cid
	PieceCid      cid.Cid
	Selector      string
	Miners        []string
	WalletAddress string
	MaxPrice      uint64
	Notifications []*ffs.NotificationConfig
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

// GetStorageAction gets an action for a JobID. If doesn't exist, returns ErrNotFound.
func (s *Store) GetStorageAction(jid ffs.JobID) (StorageAction, error) {
	var a StorageAction
	buf, err := s.ds.Get(makeStorageActionKey(jid))
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

// PutStorageAction saves a new Action for a Job.
func (s *Store) PutStorageAction(jid ffs.JobID, a StorageAction) error {
	buf, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("json marshaling: %s", err)
	}
	if err := s.ds.Put(makeStorageActionKey(jid), buf); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// GetRetrievalAction returns a the RetrievalAction corresponding to the
// RetrievalJob id.
func (s *Store) GetRetrievalAction(jid ffs.JobID) (RetrievalAction, error) {
	var a RetrievalAction
	buf, err := s.ds.Get(makeRetrievalActionKey(jid))
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

// PutRetrievalAction saves the RetrievalAction corresponding to a RetrievalJob.
func (s *Store) PutRetrievalAction(jid ffs.JobID, a RetrievalAction) error {
	buf, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("json marshaling: %s", err)
	}
	if err := s.ds.Put(makeRetrievalActionKey(jid), buf); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

func makeStorageActionKey(jid ffs.JobID) datastore.Key {
	return dsBaseStorageAction.ChildString(jid.String())
}

func makeRetrievalActionKey(jid ffs.JobID) datastore.Key {
	return dsBaseRetrievalAction.ChildString(jid.String())
}
