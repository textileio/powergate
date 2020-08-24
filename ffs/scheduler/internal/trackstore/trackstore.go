package trackstore

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
)

// Store persist information about storage configs which
// should be repaired or renewed. It always contain the
// latest Storage Config value of a Cid to execute those actions.
// This store is used by the Scheduler background jobs that
// repair or renew storage configurations.
type Store struct {
	lock        sync.Mutex
	ds          datastore.Datastore
	repairables map[cid.Cid]struct{}
	renewables  map[cid.Cid]struct{}
}

func New(ds datastore.Datastore) (*Store, error) {
	s := &Store{
		ds:          ds,
		repairables: map[cid.Cid]struct{}{},
		renewables:  map[cid.Cid]struct{}{},
	}
	if err := s.loadCaches(); err != nil {
		return nil, fmt.Errorf("loading renewable/repairable caches: %s", err)
	}
	return s, nil
}

func (s *Store) Get(c cid.Cid) (ffs.StorageConfig, error) {
	v, err := s.ds.Get(datastore.NewKey(c.String()))
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("getting storage config: %s", err)
	}
	var sc ffs.StorageConfig
	if err := json.Unmarshal(v, &sc); err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("unmarshaling storage config: %s", err)
	}
	return sc, nil
}

// Put updates the StorageConfig tracking state for a Cid.
// If the StorageConfig is repairable or renewable, it will be
// added (or updated if exist) for a Cid. If it isn't repairable
// or renewable, it will ensure it's removed from the store
// if exists. This last point happens when a StorageConfig
// which was repairable/renewable get that feature disabled.
func (s *Store) Put(c cid.Cid, sc ffs.StorageConfig) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	isRepairable := sc.Repairable
	isRenewable := sc.Cold.Enabled && sc.Cold.Filecoin.Renew.Enabled

	key := datastore.NewKey(c.String())
	if isRepairable || isRenewable {
		if isRepairable {
			s.repairables[c] = struct{}{}
		}
		if isRenewable {
			s.renewables[c] = struct{}{}
		}
		buf, err := json.Marshal(sc)
		if err != nil {
			return fmt.Errorf("marshaling storage config: %s", err)
		}
		if err := s.ds.Put(key, buf); err != nil {
			return fmt.Errorf("putting renewable/repairable storageconfig: %s", err)
		}
		return nil
	}

	// The provided Storage Config isn't interesting to
	// be tracked for renewal or repair.
	// Check if we were tracking it before, and if that's the case
	// remove it.
	_, okRenewable := s.renewables[c]
	_, okRepairable := s.repairables[c]
	if okRenewable || okRepairable {
		if err := s.ds.Delete(key); err != nil {
			return fmt.Errorf("deleting disabled storage config: %s", err)
		}
	}
	return nil

}

// Remove removes a Cid from the store, usually meaning
// the Cid storage config shouldn't be tracked for repair or
// renewal anymore.
func (s *Store) Remove(c cid.Cid) error {
	if err := s.ds.Delete(datastore.NewKey(c.String())); err != nil {
		return fmt.Errorf("deleting a tracked storage config: %s", err)
	}
	return nil
}

// GetRepairables returns the list of Cids which
// have a repairable Storage Config.
func (s *Store) GetRepairables() ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	res := make([]cid.Cid, len(s.repairables))
	i := 0
	for k := range s.repairables {
		res[i] = k
		i++
	}
	return res, nil
}

// GetRenewables returns the list of Cids which
// have a renewable Storage Config.
func (s *Store) GetRenewables() ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	res := make([]cid.Cid, len(s.renewables))
	i := 0
	for k := range s.renewables {
		res[i] = k
		i++
	}
	return res, nil
}

func (s *Store) loadCaches() error {
	q := query.Query{}
	r, err := s.ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying persisted tracked storage configs: %s", err)
	}
	defer r.Close()
	for v := range r.Next() {
		var sc ffs.StorageConfig
		if err := json.Unmarshal(v.Value, &sc); err != nil {
			return fmt.Errorf("unmarshaling storageconfig: %s", err)
		}
		c, err := cid.Decode(v.Key)
		if err != nil {
			return fmt.Errorf("decoding cid: %s", err)
		}
		if sc.Repairable {
			s.repairables[c] = struct{}{}
		}
		if sc.Cold.Enabled && sc.Cold.Filecoin.Renew.Enabled {
			s.renewables[c] = struct{}{}
		}
	}
	return nil

}
