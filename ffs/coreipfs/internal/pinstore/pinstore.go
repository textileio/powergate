package pinstore

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/ffs"
)

var (
	pinBaseKey = datastore.NewKey("pins")
	log        = logger.Logger("ffs-pinstore")
)

// Store saves information about pinned Cids per APIID.
// It can be understood as the pinset for a set of APIIDs.
// There're two types of pins: stage-pins and full-pins.
// Stage-pins indicate a form of soft-pinning that clients might
// use as an indication of unpinnable Cids by GC processes.
type Store struct {
	lock  sync.Mutex
	ds    datastore.TxnDatastore
	cache map[cid.Cid]PinnedCid
}

// PinnedCid contains information about a pinned
// Cid from multiple APIIDs.
type PinnedCid struct {
	Cid  cid.Cid
	Pins []Pin
}

// Pin describes a pin of a Cid from a APIID.
// The Stage field indicates if the pin is a stage-pin.
type Pin struct {
	APIID     ffs.APIID
	Staged    bool
	CreatedAt int64
}

// New returns a new Store.
func New(ds datastore.TxnDatastore) (*Store, error) {
	cache, err := populateCache(ds)
	if err != nil {
		return nil, fmt.Errorf("populating cache: %s", err)
	}
	return &Store{ds: ds, cache: cache}, nil
}

// AddStaged pins a Cid for APIID with a staged-pin.
// If c is already stage-pinned, its stage-pin timestamp will be refreshed.
// If c is already fully-pinned, this call is a noop (full-pin will be kept).
func (s *Store) AddStaged(iid ffs.APIID, c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var r PinnedCid
	if cr, ok := s.cache[c]; ok {
		r = cr
	} else {
		r = PinnedCid{Cid: c}
	}

	for i, p := range r.Pins {
		if p.APIID == iid {
			if !p.Staged {
				// Looks like the APIID had this Cid
				// pinned with Hot Storage, and later
				// decided to stage the Cid again.
				// Don't mark this Cid as stage-pin since
				// that would be wrong; keep the strong pin.
				// This Cid isn't GCable.
				return nil
			}
			// If the Cid is pinned because of a stage,
			// and is re-staged then simply update its
			// CreatedAt, so it will survive longer to a
			// GC.
			r.Pins[i].CreatedAt = time.Now().Unix()
			return s.persist(r)
		}
	}

	// If the Cid is not present, create it as a staged pin.
	p := Pin{
		APIID:     iid,
		Staged:    true,
		CreatedAt: time.Now().Unix(),
	}
	r.Pins = append(r.Pins, p)

	return s.persist(r)
}

// Add marks c as fully-pinned by iid.
// If c is already stage-pinned, then is switched to fully-pinned.
// If c is already fully-pinned, then only its timestamp gets refreshed.
func (s *Store) Add(iid ffs.APIID, c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var r PinnedCid
	if cr, ok := s.cache[c]; ok {
		r = cr
	} else {
		r = PinnedCid{Cid: c}
	}

	var p *Pin
	for i := range r.Pins {
		if r.Pins[i].APIID == iid {
			p = &r.Pins[i]
			if !p.Staged {
				// Log this situation since might be interesting.
				log.Warnf("%s re-pinning already pinned %s", iid, c)
			}
			break
		}
	}
	if p == nil {
		r.Pins = append(r.Pins, Pin{})
		p = &r.Pins[len(r.Pins)-1]
	}
	*p = Pin{
		APIID:     iid,
		Staged:    false,
		CreatedAt: time.Now().Unix(),
	}

	return s.persist(r)
}

// RefCount returns two integers (total, staged).
// total is the total number of ref counts for the Cid.
// staged is the total number of ref counts corresponding to
// staged pins. total includes staged, this means that:
// * total >= staged
// * non-staged pins = total - staged.
func (s *Store) RefCount(c cid.Cid) (int, int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	r, ok := s.cache[c]
	if !ok {
		return 0, 0
	}

	var stagedPins int
	for _, p := range r.Pins {
		if p.Staged {
			stagedPins++
		}
	}

	return len(r.Pins), stagedPins
}

// IsPinnedBy returns true if the Cid is pinned for APIID.
// Both strong and staged pins are considered.
func (s *Store) IsPinnedBy(iid ffs.APIID, c cid.Cid) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	r, ok := s.cache[c]
	if !ok {
		return false
	}

	for _, p := range r.Pins {
		if p.APIID == iid {
			return true
		}
	}
	return false
}

// IsPinned returns true if c is pinned by at least
// one APIID.
func (s *Store) IsPinned(c cid.Cid) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.cache[c]
	return ok
}

// Remove unpins c for iid regarding any pin type.
// If c is unpinned for iid, this is a noop.
func (s *Store) Remove(iid ffs.APIID, c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	r, ok := s.cache[c]
	if !ok {
		log.Warnf("%s removing globally unpinned cid %s", iid, c)
		return nil
	}

	c1idx := -1
	for i, p := range r.Pins {
		if p.APIID == iid {
			c1idx = i
			break
		}
	}
	if c1idx == -1 {
		log.Warnf("%s removing unpinned cid %s", iid, c)
		return nil
	}
	r.Pins[c1idx] = r.Pins[len(r.Pins)-1]
	r.Pins = r.Pins[:len(r.Pins)-1]

	return s.persist(r)
}

// RemoveStaged deletes from the pinstore c if all
// existing pins are stage-pins, if not it fails.
// This is a safe method used by GCs to unpin unpinnable cids.
func (s *Store) RemoveStaged(c cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	pc, ok := s.cache[c]
	if !ok {
		return fmt.Errorf("c isn't pinned")
	}

	for _, p := range pc.Pins {
		if !p.Staged {
			return fmt.Errorf("all pins should be stage type")
		}
	}

	if err := s.ds.Delete(makeKey(c)); err != nil {
		return fmt.Errorf("deleting from datastore: %s", err)
	}
	delete(s.cache, c)

	return nil
}

// GetAll returns all pinned cids.
func (s *Store) GetAll() ([]PinnedCid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var res []PinnedCid
	for _, v := range s.cache {
		v1 := PinnedCid{
			Cid:  v.Cid,
			Pins: make([]Pin, len(v.Pins)),
		}
		for i, p := range v.Pins {
			v1.Pins[i] = p
		}
		res = append(res, v1)
	}
	return res, nil
}

// GetAllOnlyStaged returns all cids that only have stage-pins.
func (s *Store) GetAllOnlyStaged() ([]PinnedCid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	var res []PinnedCid
Loop:
	for _, v := range s.cache {
		for _, p := range v.Pins {
			if !p.Staged {
				continue Loop
			}
		}

		res = append(res, v)
	}
	return res, nil
}

// persist persists a PinnedCid in the datastore.
func (s *Store) persist(r PinnedCid) error {
	k := makeKey(r.Cid)

	if len(r.Pins) == 0 {
		if err := s.ds.Delete(k); err != nil {
			return fmt.Errorf("delete from datastore: %s", err)
		}
		delete(s.cache, r.Cid)

		return nil
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshaling to datastore: %s", err)
	}
	if err := s.ds.Put(k, buf); err != nil {
		return fmt.Errorf("put in datastore: %s", err)
	}
	s.cache[r.Cid] = r

	return nil
}

func populateCache(ds datastore.TxnDatastore) (map[cid.Cid]PinnedCid, error) {
	q := query.Query{Prefix: pinBaseKey.String()}
	res, err := ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing populating cache query: %s", err)
		}
	}()

	ret := map[cid.Cid]PinnedCid{}
	for res := range res.Next() {
		if res.Error != nil {
			return nil, fmt.Errorf("query item result: %s", err)
		}
		var pc PinnedCid
		if err := json.Unmarshal(res.Value, &pc); err != nil {
			return nil, fmt.Errorf("unmarshaling result: %s", err)
		}
		ret[pc.Cid] = pc
	}
	return ret, nil
}

func makeKey(c cid.Cid) datastore.Key {
	return pinBaseKey.ChildString(c.String())
}
