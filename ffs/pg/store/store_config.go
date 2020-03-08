package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/fil-tools/ffs"
	"github.com/textileio/fil-tools/ffs/pg"
)

// SaveConfig persist Config information of the Powergate instance
func (cs *ConfigStore) SaveConfig(c pg.Config) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %s", err)
	}
	if err := cs.ds.Put(makeConfigKey(cs.iid), buf); err != nil {
		return fmt.Errorf("saving to store: %s", err)
	}
	return nil
}

// GetConfig gets the current configuration of the Powergate instance.
// If no configuration exist, it returns ErrConfigNotFound.
func (cs *ConfigStore) GetConfig() (*pg.Config, error) {
	buf, err := cs.ds.Get(makeConfigKey(cs.iid))
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, pg.ErrConfigNotFound
		}
		return nil, fmt.Errorf("getting config from store: %s", err)
	}
	var c pg.Config
	if err := json.Unmarshal(buf, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %s", err)
	}
	return &c, nil
}

func makeConfigKey(iid ffs.InstanceID) datastore.Key {
	return makeInstanceKey(iid).Child(dsBaseConfig)
}
