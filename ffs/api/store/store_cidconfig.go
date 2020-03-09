package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
)

// PushCidConfig saves a new desired configuration for storing a Cid
func (cs *ConfigStore) PushCidConfig(c ffs.AddAction) error {
	if !c.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cidconfig: %s", err)
	}
	if err := cs.ds.Put(makeCidConfigKey(cs.iid, c.Cid), buf); err != nil {
		return fmt.Errorf("saving cidconfig to store: %s", err)
	}
	return nil
}

// GetCidConfig returns the current desired config for the Cid
func (cs *ConfigStore) GetCidConfig(c cid.Cid) (*ffs.AddAction, error) {
	buf, err := cs.ds.Get(makeCidConfigKey(cs.iid, c))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, api.ErrConfigNotFound
		}
		return nil, err
	}
	var conf ffs.AddAction
	if err := json.Unmarshal(buf, &conf); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %s", err)
	}
	return &conf, nil
}

func makeCidConfigKey(iid ffs.InstanceID, c cid.Cid) datastore.Key {
	return dsBase.Child(dsBaseCidConfig).ChildString(c.String())
}
