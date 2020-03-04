package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/fastapi"
)

// SaveConfig persist Config information of the FastAPI instance
func (cs *ConfigStore) SaveConfig(c fastapi.Config) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %s", err)
	}
	if err := cs.ds.Put(makeConfigKey(cs.iid), buf); err != nil {
		return fmt.Errorf("saving to store: %s", err)
	}
	return nil
}

// GetConfig gets the current configuration of the FastAPI instance.
// If no configuration exist, it returns ErrConfigNotFound.
func (cs *ConfigStore) GetConfig() (*fastapi.Config, error) {
	buf, err := cs.ds.Get(makeConfigKey(cs.iid))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fastapi.ErrConfigNotFound
		}
		return nil, fmt.Errorf("getting config from store: %s", err)
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
