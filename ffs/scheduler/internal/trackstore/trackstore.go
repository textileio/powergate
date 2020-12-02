package trackstore

import (
	"encoding/json"
	"fmt"
	"strings"
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
	lock sync.Mutex
	ds   datastore.Datastore
}

// TrackedStorageConfig has information about a StorageConfig from
// an APIID which is repairable/renewable.
type TrackedStorageConfig struct {
	IID           ffs.APIID
	StorageConfig ffs.StorageConfig
}

// TrackedCid contains tracked storage configs for a Cid.
type TrackedCid struct {
	Cid     cid.Cid
	Tracked []TrackedStorageConfig
}

// New retruns a new Store.
func New(ds datastore.Datastore) (*Store, error) {
	return &Store{ds: ds}, nil
}

// Put updates the StorageConfig tracking state for a Cid.
// If the StorageConfig is repairable or renewable, it will be
// added (or updated if exist) for a Cid. If it isn't repairable
// or renewable, it will ensure it's removed from the store
// if exists. This last point happens when a StorageConfig
// which was repairable/renewable get that feature disabled.
func (s *Store) Put(iid ffs.APIID, c cid.Cid, sc ffs.StorageConfig) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	scs, err := s.get(c)
	if err != nil {
		return fmt.Errorf("getting current storage configs for cid: %s", err)
	}
	idx := -1
	for i := range scs {
		if scs[i].IID == iid {
			idx = i
			break
		}
	}

	isRepairable := sc.Repairable
	isRenewable := sc.Cold.Enabled && sc.Cold.Filecoin.Renew.Enabled
	// If has any of the interesting features, it should be in the slice.
	if isRepairable || isRenewable {
		if idx == -1 {
			nsc := TrackedStorageConfig{IID: iid, StorageConfig: sc}
			scs = append(scs, nsc)
		} else {
			scs[idx].StorageConfig = sc
		}
	} else { // Not interesting storage config to track, remove if present or simply return.
		if idx != -1 {
			scs[idx] = scs[len(scs)-1]
			scs = scs[:len(scs)-1]
		} else {
			return nil
		}
	}

	if err := s.persist(c, scs); err != nil {
		return fmt.Errorf("persisting updated storage configs for cid: %s", err)
	}

	return nil
}

// Remove removes tracking the storage config from iid.
func (s *Store) Remove(iid ffs.APIID, c cid.Cid) error {
	scs, err := s.get(c)
	if err != nil {
		return fmt.Errorf("getting current storage configs for cid: %s", err)
	}
	idx := -1
	for i := range scs {
		if scs[i].IID == iid {
			idx = i
			break
		}
	}
	if idx != -1 {
		scs[idx] = scs[len(scs)-1]
		scs = scs[:len(scs)-1]
	}
	if err := s.persist(c, scs); err != nil {
		return fmt.Errorf("persisting updated storage configs for cid: %s", err)
	}
	return nil
}

// GetRepairables returns all the tracked repairable storage configs.
func (s *Store) GetRepairables() ([]TrackedCid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tcs, err := s.query(false, true)
	if err != nil {
		return nil, fmt.Errorf("getting repairable cids from datastore: %s", err)
	}

	return tcs, nil
}

// GetRenewables returns all the tracked renewable storage configs.
func (s *Store) GetRenewables() ([]TrackedCid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tcs, err := s.query(true, false)
	if err != nil {
		return nil, fmt.Errorf("getting renewable cids from datastore: %s", err)
	}

	return tcs, nil
}

func (s *Store) query(renewable bool, repairable bool) ([]TrackedCid, error) {
	q := query.Query{}
	r, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying persisted tracked storage configs: %s", err)
	}
	defer func() { _ = r.Close() }()

	var res []TrackedCid
	for v := range r.Next() {
		if v.Error != nil {
			return nil, fmt.Errorf("iterating query result: %s", v.Error)
		}
		var tscs []TrackedStorageConfig
		if err := json.Unmarshal(v.Value, &tscs); err != nil {
			return nil, fmt.Errorf("unmarshaling storage configs: %s", err)
		}
		cidStr := strings.TrimPrefix(v.Key, "/")
		c, err := cid.Decode(cidStr)
		if err != nil {
			return nil, fmt.Errorf("decoding cid: %s", err)
		}

		tc := TrackedCid{Cid: c}
		for _, tsc := range tscs {
			isRepairable := tsc.StorageConfig.Repairable
			isRenewable := tsc.StorageConfig.Cold.Enabled && tsc.StorageConfig.Cold.Filecoin.Renew.Enabled
			if (repairable && isRepairable) || (renewable && isRenewable) {
				tc.Tracked = append(tc.Tracked, tsc)
			}
		}
		if len(tc.Tracked) > 0 {
			res = append(res, tc)
		}
	}
	return res, nil
}

func (s *Store) persist(c cid.Cid, scs []TrackedStorageConfig) error {
	// If the list of storage configs resulted to be empty, then remove it from datastore.
	key := datastore.NewKey(c.String())
	if len(scs) == 0 {
		if err := s.ds.Delete(key); err != nil {
			return fmt.Errorf("deleting empty storage configs for cid: %s", err)
		}
		return nil
	}

	// Persist the updated storage config list for the cid.
	buf, err := json.Marshal(scs)
	if err != nil {
		return fmt.Errorf("marshaling storage configs: %s", err)
	}
	if err := s.ds.Put(key, buf); err != nil {
		return fmt.Errorf("saving storage config list in datastore: %s", err)
	}
	return nil
}

func (s *Store) get(c cid.Cid) ([]TrackedStorageConfig, error) {
	v, err := s.ds.Get(datastore.NewKey(c.String()))
	if err == datastore.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting tracked storage configs: %s", err)
	}
	var tsc []TrackedStorageConfig
	if err := json.Unmarshal(v, &tsc); err != nil {
		return nil, fmt.Errorf("unmarshaling tracked storage configs: %s", err)
	}
	return tsc, nil
}
