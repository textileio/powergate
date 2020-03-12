package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
)

// PutInstanceConfig persist Config information of the Api instance
func (cs *ConfigStore) PutInstanceConfig(c api.Config) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %s", err)
	}
	if err := cs.ds.Put(makeConfigKey(cs.iid), buf); err != nil {
		return fmt.Errorf("saving to store: %s", err)
	}
	return nil
}

// GetInstanceConfig gets the current configuration of the Api instance.
// If no configuration exist, it returns ErrConfigNotFound.
func (cs *ConfigStore) GetInstanceConfig() (*api.Config, error) {
	buf, err := cs.ds.Get(makeConfigKey(cs.iid))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, api.ErrConfigNotFound
		}
		return nil, fmt.Errorf("getting config from store: %s", err)
	}
	var c api.Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %s", err)
	}
	return &c, nil
}

func makeConfigKey(iid ffs.InstanceID) datastore.Key {
	return makeInstanceKey(iid).Child(dsBaseConfig)
}
