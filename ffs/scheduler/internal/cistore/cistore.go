package cistore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("cid info not found")
)

// Store is an Datastore implementation of CidInfoStore.
type Store struct {
	ds datastore.Datastore
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Get  gets the current stored state of a Cid.
func (s *Store) Get(c cid.Cid) (ffs.CidInfo, error) {
	var ci ffs.CidInfo
	buf, err := s.ds.Get(makeKey(c))
	if err == datastore.ErrNotFound {
		return ci, ErrNotFound
	}
	if err != nil {
		return ci, fmt.Errorf("getting cid info from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &ci); err != nil {
		return ci, fmt.Errorf("unmarshaling cid info from datastore: %s", err)
	}
	return ci, nil
}

// Put saves a new storing state for a Cid.
func (s *Store) Put(ci ffs.CidInfo) error {
	if !ci.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("marshaling cid info for datastore: %s", err)
	}
	if err := s.ds.Put(makeKey(ci.Cid), buf); err != nil {
		return fmt.Errorf("put cid info in datastore: %s", err)
	}
	return nil
}

func makeKey(c cid.Cid) datastore.Key {
	return datastore.NewKey(c.String())
}
