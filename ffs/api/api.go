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
	defaultWalletType = "bls"

	log = logging.Logger("ffs-api")
)

var (
	// ErrMustOverrideConfig returned when trying to push config for storing a Cid
	// without the override flag.
	ErrMustOverrideConfig = errors.New("cid already pinned, consider using override flag")
	// ErrNotStored returned Store's items doesn't exist.
	ErrNotStored = errors.New("cid isn't stored")
)

// API is an Api instance, which owns a Lotus Address and allows to
// Store and Retrieve Cids from Hot and Cold layers.
type API struct {
	is InstanceStore
	wm ffs.WalletManager

	sched      ffs.Scheduler
	jobWatcher <-chan ffs.Job

	lock        sync.Mutex
	config      Config
	jobWatchers []jobWatcher
	ctx         context.Context
	cancel      context.CancelFunc
	finished    chan struct{}
}

type jobWatcher struct {
	jobIDs []ffs.JobID
	ch     chan ffs.Job
}

// New returns a new Api instance.
func New(ctx context.Context, iid ffs.APIID, is InstanceStore, sch ffs.Scheduler, wm ffs.WalletManager, dc ffs.DefaultCidConfig) (*API, error) {
	if err := dc.Validate(); err != nil {
		return nil, fmt.Errorf("default cid config is invalid: %s", err)
	}
	addr, err := wm.NewAddress(ctx, defaultWalletType)
	if err != nil {
		return nil, fmt.Errorf("creating new wallet addr: %s", err)
	}
	config := Config{
		ID:               iid,
		WalletAddr:       addr,
		DefaultCidConfig: dc,
	}
	ctx, cancel := context.WithCancel(context.Background())
	i := new(ctx, iid, is, wm, config, sch, cancel)
	if err := i.is.PutConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.config.ID, err)
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
		is:         is,
		wm:         wm,
		config:     config,
		sched:      sch,
		jobWatcher: sch.Watch(iid),

		cancel:   cancel,
		ctx:      ctx,
		finished: make(chan struct{}),
	}
	go i.watchJobs()
	return i
}

// ID returns the ID.
func (i *API) ID() ffs.APIID {
	return i.config.ID
}

// WalletAddr returns the Lotus wallet address.
func (i *API) WalletAddr() string {
	return i.config.WalletAddr
}

// GetDefaultCidConfig returns the default instance Cid config, prepared for a particular Cid.
func (i *API) GetDefaultCidConfig(c cid.Cid) ffs.CidConfig {
	i.lock.Lock()
	defer i.lock.Unlock()
	return newDefaultPushConfig(c, i.config.DefaultCidConfig).Config
}

// GetCidConfig returns the current CidConfig for a Cid.
func (i *API) GetCidConfig(c cid.Cid) (ffs.CidConfig, error) {
	conf, err := i.is.GetCidConfig(c)
	if err != nil {
		return ffs.CidConfig{}, fmt.Errorf("getting cid config from store: %s", err)
	}
	return conf, nil
}

// SetDefaultCidConfig sets a new default CidConfig.
func (i *API) SetDefaultCidConfig(c ffs.DefaultCidConfig) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("default cid config is invalid: %s", err)
	}
	i.config.DefaultCidConfig = c
	return nil
}

// Show returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns ErrNotStore.
func (i *API) Show(c cid.Cid) (ffs.CidInfo, error) {
	inf, err := i.sched.GetCidInfo(c)
	if err == scheduler.ErrNotFound {
		return inf, ErrNotStored
	}
	if err != nil {
		return inf, fmt.Errorf("getting cid information: %s", err)
	}
	return inf, nil
}

// Info returns instance information.
func (i *API) Info(ctx context.Context) (InstanceInfo, error) {
	pins, err := i.is.GetCids()
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("getting pins from instance: %s", err)
	}
	balance, err := i.wm.Balance(ctx, i.config.WalletAddr)
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("getting balance of %s: %s", i.config.WalletAddr, err)
	}
	return InstanceInfo{
		ID:               i.config.ID,
		DefaultCidConfig: i.config.DefaultCidConfig,
		Wallet: WalletInfo{
			Address: i.config.WalletAddr,
			Balance: balance,
		},
		Pins: pins,
	}, nil
}

// Watch subscribes to Job status changes. If jids is empty, it subscribes to
// all Job status changes corresonding to the instance. If jids is not empty,
// it immediately sends current state of those Jobs. If empty, it doesn't.
func (i *API) Watch(jids ...ffs.JobID) (<-chan ffs.Job, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	log.Info("registering watcher")

	var jobs []ffs.Job
	for _, jid := range jids {
		j, err := i.sched.GetJob(jid)
		if err == scheduler.ErrNotFound {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("getting current job state: %s", err)
		}
		jobs = append(jobs, j)
	}

	ch := make(chan ffs.Job, 1)
	i.jobWatchers = append(i.jobWatchers, jobWatcher{jobIDs: jids, ch: ch})
	for _, j := range jobs {
		select {
		case ch <- j:
		default:
			log.Warnf("dropped notifying current job state on slow receiver on %s", i.config.ID)
		}
	}

	return ch, nil
}

// Unwatch unregisters a ch returned by Watch to stop receiving updates.
func (i *API) Unwatch(ch <-chan ffs.Job) {
	i.lock.Lock()
	defer i.lock.Unlock()
	for j, w := range i.jobWatchers {
		if w.ch == ch {
			close(w.ch)
			i.jobWatchers[j] = i.jobWatchers[len(i.jobWatchers)-1]
			i.jobWatchers = i.jobWatchers[:len(i.jobWatchers)-1]
			return
		}
	}
}

// PushConfig push a new configuration for the Cid in the Hot and
// Cold layer. If WithOverride opt isn't set it errors with ErrMustOverrideConfig
func (i *API) PushConfig(c cid.Cid, opts ...PushConfigOption) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	addConfig := newDefaultPushConfig(c, i.config.DefaultCidConfig)
	for _, opt := range opts {
		if err := opt(&addConfig); err != nil {
			return ffs.EmptyJobID, fmt.Errorf("config option: %s", err)
		}
	}
	if !addConfig.OverrideConfig {
		_, err := i.is.GetCidConfig(c)
		if err == nil {
			return ffs.EmptyJobID, ErrMustOverrideConfig
		}
		if err != ErrNotFound {
			return ffs.EmptyJobID, fmt.Errorf("getting cid config: %s", err)
		}
	}
	if err := addConfig.Config.Validate(); err != nil {
		return ffs.EmptyJobID, err
	}

	conf := ffs.PushConfigAction{
		InstanceID: i.config.ID,
		Config:     addConfig.Config,
		WalletAddr: i.config.WalletAddr,
	}
	jid, err := i.sched.PushConfig(conf)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.is.PutCidConfig(conf.Config); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
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
	i.cancel()
	<-i.finished
	i.sched.Unwatch(i.jobWatcher)
	return nil
}

func (i *API) watchJobs() {
	defer close(i.finished)
	for {
		select {
		case <-i.ctx.Done():
			log.Infof("terminating job watching in %s", i.config.ID)
			return
		case j, ok := <-i.jobWatcher:
			if !ok {
				panic("scheduler closed the watching channel")
			}
			log.Info("received notification from jobstore")
			i.lock.Lock()
			log.Infof("notifying %d subscribed watchers", len(i.jobWatchers))
			for k, w := range i.jobWatchers {
				shouldNotify := len(w.jobIDs) == 0
				for _, jid := range w.jobIDs {
					if jid == j.ID {
						shouldNotify = true
						break
					}
				}
				log.Infof("evaluating watcher %d, shouldNotify %v", k, shouldNotify)
				if shouldNotify {
					select {
					case w.ch <- j:
						log.Info("notifying watcher")
					default:
						log.Warnf("skipping slow api watcher")
					}
				}
			}
			i.lock.Unlock()
		}
	}
}
