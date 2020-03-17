package pcstore

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-sched-pcstore")

	dsBase = datastore.NewKey("pcstore")
)

type Store struct {
	ds datastore.Datastore
}

var _ scheduler.PushConfigStore = (*Store)(nil)

// New returns a new JobStore backed by the Datastore.
func New(ds datastore.Datastore) *Store {
	return &Store{
		ds: ds,
	}
}

// GetCidInfo  gets the current stored state of a Cid
func (s *Store) Get(jid ffs.JobID) (ffs.PushConfigAction, error) {
	var pca ffs.PushConfigAction
	buf, err := s.ds.Get(makeKey(jid))
	if err == datastore.ErrNotFound {
		return pca, api.ErrNotFound
	}
	if err != nil {
		return pca, fmt.Errorf("get from datastore: %s", err)
	}
	if err := json.Unmarshal(buf, &pca); err != nil {
		return pca, fmt.Errorf("unmarshaling push config from datastore: %s", err)
	}
	return pca, nil
}

// PutCidInfo saves a new storing state for a Cid
func (s *Store) Put(ji ffs.JobID, pca ffs.PushConfigAction) error {
	buf, err := json.Marshal(pca)
	if err != nil {
		return fmt.Errorf("marshaling push config: %s", err)
	}
	if err := s.ds.Put(makeKey(ji), buf); err != nil {
		return fmt.Errorf("saving push config in datastore: %s", err)
	}
	return nil
}

func (s *Store) GetRenewable() ([]ffs.PushConfigAction, error) {
	q := query.Query{Prefix: dsBase.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query in datastore: %s", err)
	}
	defer res.Close()

	var actions []ffs.PushConfigAction
	for r := range res.Next() {
		var pca ffs.PushConfigAction
		if err := json.Unmarshal(r.Value, &pca); err != nil {
			return nil, fmt.Errorf("unmarshalling push config action in query: %s", err)
		}
		if pca.Config.Cold.Filecoin.Renew.Enabled {
			actions = append(actions, pca)
		}
	}
	return actions, nil
}

func makeKey(jid ffs.JobID) datastore.Key {
	return dsBase.ChildString(jid.String())
}
