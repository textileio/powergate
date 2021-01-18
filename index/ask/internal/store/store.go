package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/v2/index/ask"
)

var (
	dsKey = datastore.NewKey("index")
)

// Store persists ask index into a datastore.
type Store struct {
	ds datastore.Datastore
}

// New returns a new store for ask index.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Save persist the index into the datastore.
func (s *Store) Save(idx ask.Index) error {
	buf, err := json.Marshal(idx)
	if err != nil {
		return fmt.Errorf("marshaling new index: %s", err)
	}
	if err = s.ds.Put(dsKey, buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}
	return nil
}

// Get returns the last saved ask index. If no ask index was persisted,
// it returns an valid empty index.
func (s *Store) Get() (ask.Index, error) {
	buf, err := s.ds.Get(dsKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return ask.Index{Storage: make(map[string]ask.StorageAsk)}, nil
		}
		return ask.Index{}, err
	}
	idx := ask.Index{}
	if err = json.Unmarshal(buf, &idx); err != nil {
		return ask.Index{}, err
	}
	return idx, nil
}
