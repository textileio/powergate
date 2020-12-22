package cistore

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var (
	log = logging.Logger("ffs-sched-cistore")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("cid info not found")
)

// Store is an Datastore implementation of StorageInfoStore.
type Store struct {
	ds datastore.Datastore
	// ToDo: Build some index in here for fetching by iid and cid etc
}

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Get gets the current stored state of a Cid.
func (s *Store) Get(iid ffs.APIID, c cid.Cid) (ffs.StorageInfo, error) {
	var ci ffs.StorageInfo
	buf, err := s.ds.Get(makeKey(iid, c))
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

// Query returns a list of StorageInfo matching any provided query options.
func (s *Store) Query(iids []ffs.APIID, cids []cid.Cid) ([]ffs.StorageInfo, error) {
	APIIDFilter := make(map[ffs.APIID]struct{})
	for _, id := range iids {
		APIIDFilter[id] = struct{}{}
	}

	cidFilter := make(map[cid.Cid]struct{})
	for _, cid := range cids {
		cidFilter[cid] = struct{}{}
	}

	q := query.Query{}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying StorageInfos: %v", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()
	var ret []ffs.StorageInfo
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		if len(APIIDFilter) > 0 || len(cidFilter) > 0 {
			keyParts := strings.Split(r.Key, "/")
			if len(keyParts) != 3 {
				return nil, fmt.Errorf("parsing query result key, expected 3 parts but got %d", len(keyParts))
			}
			apiid := ffs.APIID(keyParts[1])
			cid, err := util.CidFromString(keyParts[2])
			if err != nil {
				return nil, fmt.Errorf("parsing cid from key: %v", err)
			}
			if len(APIIDFilter) > 0 {
				if _, ok := APIIDFilter[apiid]; !ok {
					continue
				}
			}
			if len(cidFilter) > 0 {
				if _, ok := cidFilter[cid]; !ok {
					continue
				}
			}
		}
		var si ffs.StorageInfo
		if err := json.Unmarshal(r.Value, &si); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, si)
	}
	sort.Slice(ret, func(i, j int) bool {
		// Descending.
		return ret[i].Created.After(ret[j].Created)
	})
	return ret, nil
}

// Put saves a new storage state for a Cid.
func (s *Store) Put(ci ffs.StorageInfo) error {
	if !ci.APIID.Valid() {
		return fmt.Errorf("instance id is invalid")
	}
	if !ci.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(ci)
	if err != nil {
		return fmt.Errorf("marshaling cid info for datastore: %s", err)
	}
	if err := s.ds.Put(makeKey(ci.APIID, ci.Cid), buf); err != nil {
		return fmt.Errorf("put cid info in datastore: %s", err)
	}
	return nil
}

func makeKey(iid ffs.APIID, c cid.Cid) datastore.Key {
	return datastore.NewKey(iid.String()).ChildString(util.CidToString(c))
}
