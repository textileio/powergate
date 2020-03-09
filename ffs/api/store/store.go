package store

import (
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
)

var (
	dsBase          = datastore.NewKey("instance")
	dsBaseConfig    = datastore.NewKey("config")
	dsBaseCidConfig = datastore.NewKey("cidconfig")
	dsBaseCidInfo   = datastore.NewKey("cidinfo")
)

// ConfigStore is an implementation of api.ConfigStore interface
type ConfigStore struct {
	lock sync.Mutex
	ds   datastore.Datastore
	iid  ffs.InstanceID
}

var _ api.ConfigStore = (*ConfigStore)(nil)

// New returns a new ConfigStore
func New(iid ffs.InstanceID, ds datastore.Datastore) *ConfigStore {
	return &ConfigStore{
		iid: iid,
		ds:  ds,
	}
}

func makeInstanceKey(iid ffs.InstanceID) datastore.Key {
	return dsBase.ChildString(iid.String())
}
