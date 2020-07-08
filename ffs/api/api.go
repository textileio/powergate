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
func New(ctx context.Context, ds datastore.Datastore, iid ffs.APIID, sch *scheduler.Scheduler, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager, dc ffs.DefaultConfig) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))

	addr, err := wm.NewAddress(ctx, defaultAddressType)
	if err != nil {
		return nil, fmt.Errorf("creating new wallet addr: %s", err)
	}

	dc.Cold.Filecoin.Addr = addr

	if err := dc.Validate(); err != nil {
		return nil, fmt.Errorf("default cid config is invalid: %s", err)
	}

	addrInfo := AddrInfo{
		Name: "Initial Address",
		Addr: addr,
		Type: defaultAddressType,
	}

	config := Config{
		ID:            iid,
		Addrs:         map[string]AddrInfo{addr: addrInfo},
		DefaultConfig: dc,
	}

	ctx, cancel := context.WithCancel(context.Background())
	i := new(ctx, is, wm, pm, drm, config, sch, cancel)
	if err := i.is.putConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.cfg.ID, err)
	}
	return i, nil
}

// Load loads a saved Api instance from its ConfigStore.
func Load(ds datastore.Datastore, iid ffs.APIID, sched *scheduler.Scheduler, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager) (*API, error) {
	is := newInstanceStore(namespace.Wrap(ds, datastore.NewKey("istore")))
	c, err := is.getConfig()
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

// DefaultConfig returns the DefaultConfig.
func (i *API) DefaultConfig() ffs.DefaultConfig {
	return i.cfg.DefaultConfig
}

// GetDefaultCidConfig returns the default instance Cid config, prepared for a particular Cid.
func (i *API) GetDefaultCidConfig(c cid.Cid) ffs.CidConfig {
	i.lock.Lock()
	defer i.lock.Unlock()
	return newDefaultPushConfig(c, i.cfg.DefaultConfig).Config
}

// GetCidConfig returns the current CidConfig for a Cid.
func (i *API) GetCidConfig(c cid.Cid) (ffs.CidConfig, error) {
	conf, err := i.is.getCidConfig(c)
	if err == ErrNotFound {
		return ffs.CidConfig{}, err
	}
	if err != nil {
		return ffs.CidConfig{}, fmt.Errorf("getting cid config from store: %s", err)
	}
	return conf, nil
}

// SetDefaultConfig sets a new default CidConfig.
func (i *API) SetDefaultConfig(c ffs.DefaultConfig) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("default cid config is invalid: %s", err)
	}
	i.cfg.DefaultConfig = c
	return i.is.putConfig(i.cfg)
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
		ID:            i.cfg.ID,
		DefaultConfig: i.cfg.DefaultConfig,
		Balances:      balances,
		Pins:          pins,
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
