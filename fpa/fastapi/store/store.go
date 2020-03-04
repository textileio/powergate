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

type ConfigStore struct {
	lock sync.Mutex
	ds   datastore.Datastore
	iid  fpa.InstanceID
}

var _ fastapi.ConfigStore = (*ConfigStore)(nil)

func New(iid fpa.InstanceID, ds datastore.Datastore) *ConfigStore {
	return &ConfigStore{
		iid: iid,
		ds:  ds,
	}
}
