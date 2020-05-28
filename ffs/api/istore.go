package api

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
)

var (
	dsDefaultConfig = datastore.NewKey("defaultconfig")
	dsCidConfig     = datastore.NewKey("cidconfig")
)

type instanceStore struct {
	lock sync.Mutex
	ds   datastore.Datastore
}

func newInstanceStore(ds datastore.Datastore) *instanceStore {
	return &instanceStore{
		ds: ds,
	}
}

func (s *instanceStore) putConfig(c Config) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling instance config to datastore: %s", err)
	}
	if err := s.ds.Put(dsDefaultConfig, buf); err != nil {
		return fmt.Errorf("saving to datastore: %s", err)
	}
	return nil
}

func (s *instanceStore) getConfig() (Config, error) {
	buf, err := s.ds.Get(dsDefaultConfig)
	if err != nil {
		if err == datastore.ErrNotFound {
			return Config{}, ErrNotFound
		}
		return Config{}, fmt.Errorf("getting instance config from datastore: %s", err)
	}
	var c Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return Config{}, fmt.Errorf("unmarshaling config from datastore: %s", err)
	}
	return c, nil
}

func (s *instanceStore) putCidConfig(c ffs.CidConfig) error {
	if !c.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cid config to datastore: %s", err)
	}
	if err := s.ds.Put(makeCidConfigKey(c.Cid), buf); err != nil {
		return fmt.Errorf("saving cid config to datastore: %s", err)
	}
	return nil
}

func (s *instanceStore) removeCidConfig(c cid.Cid) error {
	if !c.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	if err := s.ds.Delete(makeCidConfigKey(c)); err != nil {
		return fmt.Errorf("removing from datastore: %s", err)
	}
	return nil
}

func (s *instanceStore) getCidConfig(c cid.Cid) (ffs.CidConfig, error) {
	buf, err := s.ds.Get(makeCidConfigKey(c))
	if err != nil {
		if err == datastore.ErrNotFound {
			return ffs.CidConfig{}, ErrNotFound
		}
		return ffs.CidConfig{}, err
	}
	var conf ffs.CidConfig
	if err := json.Unmarshal(buf, &conf); err != nil {
		return ffs.CidConfig{}, fmt.Errorf("unmarshaling cid config from datastore: %s", err)
	}
	return conf, nil
}

func (s *instanceStore) getCids() ([]cid.Cid, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{
		Prefix:   dsCidConfig.String(),
		KeysOnly: true,
	}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying for all cids in instance: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

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

func makeCidConfigKey(c cid.Cid) datastore.Key {
	return dsCidConfig.ChildString(c.String())
}
