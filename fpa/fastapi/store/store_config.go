package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/fastapi"
)

func (cs *ConfigStore) SaveConfig(c fastapi.Config) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %s", err)
	}
	if err := cs.ds.Put(makeConfigKey(cs.iid), buf); err != nil {
		return fmt.Errorf("saving to store: %s", err)
	}
	return nil
}

func (cs *ConfigStore) GetConfig() (*fastapi.Config, error) {
	buf, err := cs.ds.Get(makeConfigKey(cs.iid))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fastapi.ErrConfigNotFound
		}
		return nil, err
	}
	var c fastapi.Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %s", err)
	}
	return &c, nil
}

func makeConfigKey(iid fpa.InstanceID) datastore.Key {
	return dsBase.Child(dsBaseConfig)
}
