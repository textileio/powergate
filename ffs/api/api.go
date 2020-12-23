package api

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	log = logging.Logger("ffs-api")
)

var (
	// ErrMustOverrideConfig returned when trying to push config for storing a Cid
	// without the override flag.
	ErrMustOverrideConfig = errors.New("cid already pinned, consider using override flag")
	// ErrReplacedCidNotFound returns when replacing a Cid that isn't stored.
	ErrReplacedCidNotFound = errors.New("provided replaced cid wasn't found")
	// ErrActiveInStorage returns when a Cid is trying to be removed but still defined as active
	// on Hot or Cold storage.
	ErrActiveInStorage = errors.New("can't remove Cid, disable from Hot and Cold storage")
	// ErrHotStorageDisabled returned when trying to fetch a Cid when disabled on Hot Storage.
	// To retrieve the data, is necessary to call unfreeze by enabling the Enabled flag in
	// hot storage for that Cid.
	ErrHotStorageDisabled = errors.New("cid disabled in hot storage")
)

// API is an Api instance, which owns a Lotus Address and allows to
// Store and Retrieve Cids from hot and cold storage.
type API struct {
	is  *instanceStore
	wm  ffs.WalletManager
	drm ffs.DealRecordsManager

	sched *scheduler.Scheduler

	lock   sync.Mutex
	closed bool
	cfg    InstanceConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a new Api instance.
func New(ds datastore.Datastore, iid ffs.APIID, sch *scheduler.Scheduler, wm ffs.WalletManager, drm ffs.DealRecordsManager, dc ffs.StorageConfig, addrInfo AddrInfo) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))

	dc.Cold.Filecoin.Addr = addrInfo.Addr

	if err := dc.Validate(); err != nil {
		return nil, fmt.Errorf("default storage config is invalid: %s", err)
	}

	config := InstanceConfig{
		ID:                   iid,
		Addrs:                map[string]AddrInfo{addrInfo.Addr: addrInfo},
		DefaultStorageConfig: dc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	i := new(ctx, is, wm, drm, config, sch, cancel)
	if err := i.is.putInstanceConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.cfg.ID, err)
	}
	return i, nil
}

// Load loads a saved Api instance from its ConfigStore.
func Load(ds datastore.Datastore, iid ffs.APIID, sched *scheduler.Scheduler, wm ffs.WalletManager, drm ffs.DealRecordsManager) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))
	c, err := is.getInstanceConfig()
	if err != nil {
		return nil, fmt.Errorf("loading instance: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return new(ctx, is, wm, drm, c, sched, cancel), nil
}

func new(ctx context.Context, is *instanceStore, wm ffs.WalletManager, drm ffs.DealRecordsManager, config InstanceConfig, sch *scheduler.Scheduler, cancel context.CancelFunc) *API {
	i := &API{
		is:     is,
		wm:     wm,
		drm:    drm,
		cfg:    config,
		sched:  sch,
		cancel: cancel,
		ctx:    ctx,
	}
	return i
}

// ID returns the ID.
func (i *API) ID() ffs.APIID {
	return i.cfg.ID
}

// DefaultStorageConfig returns the default StorageConfig.
func (i *API) DefaultStorageConfig() ffs.StorageConfig {
	return i.cfg.DefaultStorageConfig
}

// SetDefaultStorageConfig sets a new default StorageConfig.
func (i *API) SetDefaultStorageConfig(c ffs.StorageConfig) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("default cid config is invalid: %s", err)
	}
	i.cfg.DefaultStorageConfig = c
	return i.is.putInstanceConfig(i.cfg)
}

// GetStorageConfigs returns the current StorageConfigs for a FFS instance, filtered by cids, if provided.
func (i *API) GetStorageConfigs(cids ...cid.Cid) (map[cid.Cid]ffs.StorageConfig, error) {
	configs, err := i.is.getStorageConfigs(cids...)
	if err == ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("getting cid configs from store: %s", err)
	}
	return configs, nil
}

// Close terminates the running Api.
func (i *API) Close() error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if i.closed {
		return nil
	}
	i.cancel()
	i.closed = true
	return nil
}
