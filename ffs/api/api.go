package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	defaultAddressType = "bls"

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
)

// API is an Api instance, which owns a Lotus Address and allows to
// Store and Retrieve Cids from Hot and Cold layers.
type API struct {
	iid ffs.APIID
	is  InstanceStore
	wm  ffs.WalletManager

	sched ffs.Scheduler

	lock   sync.Mutex
	closed bool
	cfg    Config
	ctx    context.Context
	cancel context.CancelFunc
}

// New returns a new Api instance.
func New(ctx context.Context, iid ffs.APIID, is InstanceStore, sch ffs.Scheduler, wm ffs.WalletManager, dc ffs.DefaultConfig) (*API, error) {
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
	i := new(ctx, iid, is, wm, config, sch, cancel)
	if err := i.is.PutConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.cfg.ID, err)
	}
	return i, nil
}

// Load loads a saved Api instance from its ConfigStore.
func Load(iid ffs.APIID, is InstanceStore, sched ffs.Scheduler, wm ffs.WalletManager) (*API, error) {
	c, err := is.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("loading instance: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return new(ctx, iid, is, wm, c, sched, cancel), nil
}

func new(ctx context.Context, iid ffs.APIID, is InstanceStore, wm ffs.WalletManager, config Config, sch ffs.Scheduler, cancel context.CancelFunc) *API {
	i := &API{
		is:     is,
		wm:     wm,
		cfg:    config,
		sched:  sch,
		cancel: cancel,
		ctx:    ctx,
		iid:    iid,
	}
	return i
}

// ID returns the ID.
func (i *API) ID() ffs.APIID {
	return i.cfg.ID
}

// Addrs returns the wallet addresses.
func (i *API) Addrs() []AddrInfo {
	i.lock.Lock()
	defer i.lock.Unlock()
	var addrs []AddrInfo
	for _, addr := range i.cfg.Addrs {
		addrs = append(addrs, addr)
	}
	return addrs
}

// DefaultConfig returns the DefaultConfig
func (i *API) DefaultConfig() ffs.DefaultConfig {
	return i.cfg.DefaultConfig
}

// NewAddr creates a new address managed by the FFS instance
func (i *API) NewAddr(ctx context.Context, name string, options ...NewAddressOption) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	conf := &NewAddressConfig{
		makeDefault: false,
		addressType: defaultAddressType,
	}
	for _, option := range options {
		option(conf)
	}

	exists := false
	for _, addr := range i.cfg.Addrs {
		if addr.Name == name {
			exists = true
			break
		}
	}

	if exists {
		return "", fmt.Errorf("address with name %s already exists", name)
	}

	addr, err := i.wm.NewAddress(ctx, conf.addressType)
	if err != nil {
		return "", fmt.Errorf("creating new wallet addr: %s", err)
	}
	i.cfg.Addrs[addr] = AddrInfo{
		Name: name,
		Addr: addr,
		Type: conf.addressType,
	}
	if conf.makeDefault {
		i.cfg.DefaultConfig.Cold.Filecoin.Addr = addr
	}
	if err := i.is.PutConfig(i.cfg); err != nil {
		return "", err
	}
	return addr, nil
}

// GetDefaultCidConfig returns the default instance Cid config, prepared for a particular Cid.
func (i *API) GetDefaultCidConfig(c cid.Cid) ffs.CidConfig {
	i.lock.Lock()
	defer i.lock.Unlock()
	return newDefaultPushConfig(c, i.cfg.DefaultConfig).Config
}

// GetCidConfig returns the current CidConfig for a Cid.
func (i *API) GetCidConfig(c cid.Cid) (ffs.CidConfig, error) {
	conf, err := i.is.GetCidConfig(c)
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
	return i.is.PutConfig(i.cfg)
}

// Show returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns ErrNotFound.
func (i *API) Show(c cid.Cid) (ffs.CidInfo, error) {
	inf, err := i.sched.GetCidInfo(c)
	if err == scheduler.ErrNotFound {
		return inf, ErrNotFound
	}
	if err != nil {
		return inf, fmt.Errorf("getting cid information: %s", err)
	}
	return inf, nil
}

// Info returns instance information.
func (i *API) Info(ctx context.Context) (InstanceInfo, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	pins, err := i.is.GetCids()
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

// WatchJobs subscribes to Job status changes. If jids is empty, it subscribes to
// all Job status changes corresonding to the instance. If jids is not empty,
// it immediately sends current state of those Jobs. If empty, it doesn't.
func (i *API) WatchJobs(ctx context.Context, c chan<- ffs.Job, jids ...ffs.JobID) error {
	var jobs []ffs.Job
	for _, jid := range jids {
		j, err := i.sched.GetJob(jid)
		if err == scheduler.ErrNotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("getting current job state: %s", err)
		}
		jobs = append(jobs, j)
	}

	ch := make(chan ffs.Job, 1)
	for _, j := range jobs {
		select {
		case ch <- j:
		default:
			log.Warnf("dropped notifying current job state on slow receiver on %s", i.cfg.ID)
		}
	}
	var err error
	go func() {
		err = i.sched.WatchJobs(ctx, ch, i.iid)
		close(ch)
	}()
	for j := range ch {
		if len(jids) == 0 {
			c <- j
		}
	JidLoop:
		for _, jid := range jids {
			if jid == j.ID {
				c <- j
				break JidLoop
			}
		}
	}
	if err != nil {
		return fmt.Errorf("scheduler listener: %s", err)
	}

	return nil
}

// Replace pushes a CidConfig of c2 equal to c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
func (i *API) Replace(c1 cid.Cid, c2 cid.Cid) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg, err := i.is.GetCidConfig(c1)
	if err == ErrNotFound {
		return ffs.EmptyJobID, ErrReplacedCidNotFound
	}
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("getting replaced cid config: %s", err)
	}
	cfg.Cid = c2

	if err := i.ensureValidColdCfg(cfg.Cold); err != nil {
		return ffs.EmptyJobID, err
	}

	jid, err := i.sched.PushReplace(i.iid, cfg, c1)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling replacement %s to %s: %s", c1, c2, err)
	}
	if err := i.is.PutCidConfig(cfg); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c2, err)
	}
	if err := i.is.RemoveCidConfig(c1); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return jid, nil
}

// PushConfig push a new configuration for the Cid in the Hot and
// Cold layer. If WithOverride opt isn't set it errors with ErrMustOverrideConfig
func (i *API) PushConfig(c cid.Cid, opts ...PushConfigOption) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg := newDefaultPushConfig(c, i.cfg.DefaultConfig)
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return ffs.EmptyJobID, fmt.Errorf("config option: %s", err)
		}
	}
	if !cfg.OverrideConfig {
		_, err := i.is.GetCidConfig(c)
		if err == nil {
			return ffs.EmptyJobID, ErrMustOverrideConfig
		}
		if err != ErrNotFound {
			return ffs.EmptyJobID, fmt.Errorf("getting cid config: %s", err)
		}
	}
	if err := cfg.Config.Validate(); err != nil {
		return ffs.EmptyJobID, err
	}
	if err := i.ensureValidColdCfg(cfg.Config.Cold); err != nil {
		return ffs.EmptyJobID, err
	}

	jid, err := i.sched.PushConfig(i.cfg.ID, cfg.Config)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.is.PutCidConfig(cfg.Config); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (i *API) Remove(c cid.Cid) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg, err := i.is.GetCidConfig(c)
	if err == ErrNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("getting cid config from store: %s", err)
	}
	if cfg.Hot.Enabled || cfg.Cold.Enabled {
		return ErrActiveInStorage
	}
	if err := i.sched.Untrack(c); err != nil {
		return fmt.Errorf("untracking from scheduler: %s", err)
	}
	if err := i.is.RemoveCidConfig(c); err != nil {
		return fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return nil
}

// Get returns an io.Reader for reading a stored Cid from the Hot Storage.
func (i *API) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	if !c.Defined() {
		return nil, fmt.Errorf("cid is undefined")
	}
	conf, err := i.is.GetCidConfig(c)
	if err != nil {
		return nil, fmt.Errorf("getting cid config: %s", err)
	}
	if !conf.Hot.Enabled {
		return nil, ffs.ErrHotStorageDisabled
	}
	r, err := i.sched.GetCidFromHot(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting from hot layer %s: %s", c, err)
	}
	return r, nil
}

// WatchLogs pushes human-friendly messages about Cid executions. The method is blocking
// and will continue to send messages until the context is canceled.
func (i *API) WatchLogs(ctx context.Context, ch chan<- ffs.LogEntry, c cid.Cid, opts ...GetLogsOption) error {
	_, err := i.is.GetCidConfig(c)
	if err == ErrNotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("validating cid: %s", err)
	}

	config := &GetLogsConfig{}
	for _, o := range opts {
		o(config)
	}

	ic := make(chan ffs.LogEntry)
	go func() {
		err = i.sched.WatchLogs(ctx, ic)
		close(ic)
	}()
	for le := range ic {
		if c == le.Cid && (config.jid == ffs.EmptyJobID || config.jid == le.Jid) {
			ch <- le
		}
	}
	if err != nil {
		return fmt.Errorf("listening to cid logs: %s", err)
	}

	return nil
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

func (i *API) ensureValidColdCfg(cfg ffs.ColdConfig) error {
	if cfg.Enabled && !i.isManagedAddress(cfg.Filecoin.Addr) {
		return fmt.Errorf("%v is not managed by ffs instance", cfg.Filecoin.Addr)
	}
	return nil
}

func (i *API) isManagedAddress(addr string) bool {
	for managedAddr := range i.cfg.Addrs {
		if managedAddr == addr {
			return true
		}
	}
	return false
}
