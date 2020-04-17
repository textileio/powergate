package astore

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sched-astore")
)

// Store is a Datastore backed implementation of ActionStore, which saves latests
// PushConfig actions for a Cid.
type Store struct {
	ds datastore.Datastore
}

var _ scheduler.ActionStore = (*Store)(nil)

// New returns a new ActionStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Get gets the lastest pushed Action of a Cid.
func (s *Store) Get(jid ffs.JobID) (scheduler.Action, error) {
	var a scheduler.Action
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return a, api.ErrNotFound
	}
	if err != nil {
		return a, fmt.Errorf("get from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &a); err != nil {
		return a, fmt.Errorf("unmarshaling from datastore: %s", err)
	}
	return a, nil
}

// Put saves a new Action for a Cid.
func (s *Store) Put(ji ffs.JobID, a scheduler.Action) error {
	buf, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("json marshaling: %s", err)
	}
	if err := s.ds.Put(makeKey(ji), buf); err != nil {
		return fmt.Errorf("saving in datastore: %s", err)
	}
	return nil
}

// Remove removes any Action associated with a Cid.
func (s *Store) Remove(c cid.Cid) error {
	// ToDo: if this becomes a bottleneck, consider including
	// Cid in key or make an index.
	q := query.Query{Prefix: ""}
	res, err := s.ds.Query(q)
	if err != nil {
		return fmt.Errorf("executing query in datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	for r := range res.Next() {
		var a scheduler.Action
		if err := json.Unmarshal(r.Value, &a); err != nil {
			return fmt.Errorf("unmarshalling push config action in query: %s", err)
		}
		if a.Cfg.Cid == c {
			if err := s.ds.Delete(datastore.NewKey(r.Key)); err != nil {
				return fmt.Errorf("deleting from datastore: %s", err)
			}
			return nil
		}
	}
	return scheduler.ErrNotFound

}

// GetRenewable returns all Actions that have CidConfigs that have the Renew flag enabled
// and should be inspected for Deal renewals.
func (s *Store) GetRenewable() ([]scheduler.Action, error) {
	q := query.Query{Prefix: ""}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query in datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var as []scheduler.Action
	for r := range res.Next() {
		var a scheduler.Action
		if err := json.Unmarshal(r.Value, &a); err != nil {
			return nil, fmt.Errorf("unmarshalling push config action in query: %s", err)
		}
		if a.Cfg.Cold.Enabled && a.Cfg.Cold.Filecoin.Renew.Enabled {
			as = append(as, a)
		}
	}
	return as, nil
}

func makeKey(jid ffs.JobID) datastore.Key {
	return datastore.NewKey(jid.String())
}
