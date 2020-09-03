package ristore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound indicates there's not information for the retrieval.
	ErrNotFound = errors.New("retrieval info not found")
)

// Store stores retrieval information.
type Store struct {
	ds datastore.Datastore
}

// New returns a new retrieval information store.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Get gets information about an executed retrieval.
func (s *Store) Get(rid ffs.RetrievalID) (ffs.RetrievalInfo, error) {
	var ri ffs.RetrievalInfo
	buf, err := s.ds.Get(makeKey(rid))
	if err == datastore.ErrNotFound {
		return ri, ErrNotFound
	}
	if err != nil {
		return ri, fmt.Errorf("getting retrieval info from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &ri); err != nil {
		return ri, fmt.Errorf("unmarshaling retrieval info from datastore: %s", err)
	}
	return ri, nil
}

// Put saves new information about a retrieval.
func (s *Store) Put(ri ffs.RetrievalInfo) error {
	if ri.ID == ffs.EmptyRetrievalID {
		return fmt.Errorf("retrieval id can't be empty")
	}
	buf, err := json.Marshal(ri)
	if err != nil {
		return fmt.Errorf("marshaling retrieval info for datastore: %s", err)
	}
	if err := s.ds.Put(makeKey(ri.ID), buf); err != nil {
		return fmt.Errorf("putting retrieval info in datastore: %s", err)
	}
	return nil
}

func makeKey(rid ffs.RetrievalID) datastore.Key {
	return datastore.NewKey(rid.String())
}
