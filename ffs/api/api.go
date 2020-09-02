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
	// the Hot Storage for that Cid.
	ErrHotStorageDisabled = errors.New("cid disabled in hot storage")
)

// API is an Api instance, which owns a Lotus Address and allows to
// Store and Retrieve Cids from Hot and Cold layers.
type API struct {
	is  *instanceStore
	wm  ffs.WalletManager
	pm  ffs.PaychManager
	drm ffs.DealRecordsManager

	sched *scheduler.Scheduler

	lock   sync.Mutex
	closed bool
	cfg    Config
	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a new Api instance.
func New(ds datastore.Datastore, iid ffs.APIID, sch *scheduler.Scheduler, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager, dc ffs.StorageConfig, addrInfo AddrInfo) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))

	dc.Cold.Filecoin.Addr = addrInfo.Addr

	if err := dc.Validate(); err != nil {
		return nil, fmt.Errorf("default storage config is invalid: %s", err)
	}

	config := Config{
		ID:                   iid,
		Addrs:                map[string]AddrInfo{addrInfo.Addr: addrInfo},
		DefaultStorageConfig: dc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	i := new(ctx, is, wm, pm, drm, config, sch, cancel)
	if err := i.is.putInstanceConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.cfg.ID, err)
	}
	return i, nil
}

// Load loads a saved Api instance from its ConfigStore.
func Load(ds datastore.Datastore, iid ffs.APIID, sched *scheduler.Scheduler, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))
	c, err := is.getInstanceConfig()
	if err != nil {
		return nil, fmt.Errorf("loading instance: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return new(ctx, is, wm, pm, drm, c, sched, cancel), nil
}

func new(ctx context.Context, is *instanceStore, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager, config Config, sch *scheduler.Scheduler, cancel context.CancelFunc) *API {
	i := &API{
		is:     is,
		wm:     wm,
		pm:     pm,
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

// GetStorageConfig returns the current StorageConfig for a Cid.
func (i *API) GetStorageConfig(c cid.Cid) (ffs.StorageConfig, error) {
	conf, err := i.is.getStorageConfig(c)
	if err == ErrNotFound {
		return ffs.StorageConfig{}, err
	}
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("getting cid config from store: %s", err)
	}
	return conf, nil
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

// Info returns instance information.
func (i *API) Info(ctx context.Context) (InstanceInfo, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	pins, err := i.is.getCids()
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("getting pins from instance: %s", err)
	}

	var balances []BalanceInfo
	for _, addr := range i.cfg.Addrs {
		balance, err := i.wm.Balance(ctx, addr.Addr)
		if err != nil {
			return InstanceInfo{}, fmt.Errorf("getting balance of %s: %s", addr.Addr, err)
		}
		info := BalanceInfo{
			AddrInfo: addr,
			Balance:  balance,
		}
		balances = append(balances, info)
	}

	return InstanceInfo{
		ID:                   i.cfg.ID,
		DefaultStorageConfig: i.cfg.DefaultStorageConfig,
		Balances:             balances,
		Pins:                 pins,
	}, nil
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
