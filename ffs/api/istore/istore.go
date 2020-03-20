package istore

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
)

var (
	dsBase           = datastore.NewKey("instance")
	dsInstanceConfig = datastore.NewKey("config")
	dsCidConfig      = datastore.NewKey("cidconfig")
)

// Store is an implementation of api.ConfigStore interface
type Store struct {
	lock sync.Mutex
	ds   datastore.Datastore
	iid  ffs.ApiID
}

var _ api.InstanceStore = (*Store)(nil)

// New returns a new ConfigStore
func New(iid ffs.ApiID, ds datastore.Datastore) *Store {
	return &Store{
		iid: iid,
		ds:  ds,
	}
}

// PutConfig persist Config information of the Api instance.
func (s *Store) PutConfig(c api.Config) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling instance config to datastore: %s", err)
	}
	if err := s.ds.Put(makeConfigKey(s.iid), buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}
	return nil
}

// GetConfig gets the current configuration of the Api instance.
// If no configuration exist, it returns ErrNotFound.
func (s *Store) GetConfig() (api.Config, error) {
	buf, err := s.ds.Get(makeConfigKey(s.iid))
	if err != nil {
		if err == datastore.ErrNotFound {
			return api.Config{}, api.ErrNotFound
		}
		return api.Config{}, fmt.Errorf("getting instance config from datastore: %s", err)
	}
	var c api.Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return api.Config{}, fmt.Errorf("unmarshaling config from datastore: %s", err)
	}
	return c, nil
}

// PutCidConfig saves a new desired configuration for storing a Cid.
func (s *Store) PutCidConfig(c ffs.CidConfig) error {
	if !c.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cid config to datastore: %s", err)
	}
	if err := s.ds.Put(makeCidConfigKey(s.iid, c.Cid), buf); err != nil {
		return fmt.Errorf("saving cid config to datastore: %s", err)
	}
	return nil
}

// GetCidConfig returns the current desired config for the Cid
func (s *Store) GetCidConfig(c cid.Cid) (ffs.CidConfig, error) {
	buf, err := s.ds.Get(makeCidConfigKey(s.iid, c))
	if err != nil {
		if err == datastore.ErrNotFound {
			return ffs.CidConfig{}, api.ErrNotFound
		}
		return ffs.CidConfig{}, err
	}
	var conf ffs.CidConfig
	if err := json.Unmarshal(buf, &conf); err != nil {
		return ffs.CidConfig{}, fmt.Errorf("unmarshaling cid config from datastore: %s", err)
	}
	return conf, nil
}

// GetCids returns a slice of Cids which have storing state
func (s *Store) GetCids() ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{
		Prefix:   makeInstanceKey(s.iid).Child(dsCidConfig).String(),
		KeysOnly: true,
	}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying for all cids in instance: %s", err)
	}
	defer res.Close()

	var cids []cid.Cid
	for r := range res.Next() {
		strCid := datastore.RawKey(r.Key).Name()
		c, err := cid.Decode(strCid)
		if err != nil {
			return nil, fmt.Errorf("decoding cid: %s", err)
		}
		cids = append(cids, c)
	}
	return cids, nil
}

func makeCidConfigKey(iid ffs.ApiID, c cid.Cid) datastore.Key {
	return makeInstanceKey(iid).Child(dsCidConfig).ChildString(c.String())
}

func makeConfigKey(iid ffs.ApiID) datastore.Key {
	return makeInstanceKey(iid).Child(dsInstanceConfig)
}

func makeInstanceKey(iid ffs.ApiID) datastore.Key {
	return dsBase.ChildString(iid.String())
}
