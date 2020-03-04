package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/fastapi"
)

func (cs *ConfigStore) PushCidConfig(c fpa.CidConfig) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cid config: %s", err)
	}
	if err := cs.ds.Put(makeCidConfigKey(cs.iid, c.Cid), buf); err != nil {
		return fmt.Errorf("saving cid config to store: %s", err)
	}
	return nil
}

func (cs *ConfigStore) GetCidConfig(c cid.Cid) (*fpa.CidConfig, error) {
	buf, err := cs.ds.Get(makeCidConfigKey(cs.iid, c))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fastapi.ErrConfigNotFound
		}
		return nil, err
	}
	var conf fpa.CidConfig
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %s", err)
	}
	return &conf, nil
}

func makeCidConfigKey(iid fpa.InstanceID, c cid.Cid) datastore.Key {
	return dsBase.Child(dsBaseCidConfig).ChildString(c.String())
}
