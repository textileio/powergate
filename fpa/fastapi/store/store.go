package store

import (
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/fastapi"
)

var (
	dsBase          = datastore.NewKey("instance")
	dsBaseConfig    = datastore.NewKey("config")
	dsBaseCidConfig = datastore.NewKey("cidconfig")
	dsBaseCidInfo   = datastore.NewKey("cidinfo")
)

// ConfigStore is an implementation of fastapi.ConfigStore interface
type ConfigStore struct {
	lock sync.Mutex
	ds   datastore.Datastore
	iid  fpa.InstanceID
}

var _ fastapi.ConfigStore = (*ConfigStore)(nil)

// New returns a new ConfigStore
func New(iid fpa.InstanceID, ds datastore.Datastore) *ConfigStore {
	return &ConfigStore{
		iid: iid,
		ds:  ds,
	}
}

func makeInstanceKey(iid fpa.InstanceID) datastore.Key {
	return dsBase.ChildString(iid.String())
}
