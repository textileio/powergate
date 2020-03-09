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
)

var (
	defaultWalletType = "bls"

	log = logging.Logger("ffs-api")
)

var (
	// ErrAlreaadyPinned returned when trying to push an initial config
	// for storing a Cid.
	ErrAlreadyPinned = errors.New("cid already pinned")
	// ErNotStored returned when there isn't CidInfo for a Cid
	ErrNotStored = errors.New("cid isn't stored")
)

// Instance is an Api instance, which owns a Lotus Address and allows to
// Store and Retrieve Cids from Hot and Cold layers.
type Instance struct {
	store ConfigStore
	wm    ffs.WalletManager

	sched   ffs.Scheduler
	chSched <-chan ffs.Job

	lock     sync.Mutex
	config   Config
	watchers []watcher
	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

type watcher struct {
	jobIDs []ffs.JobID
	ch     chan ffs.Job
}

// New returns a new Api instance.
func New(ctx context.Context, iid ffs.InstanceID, confstore ConfigStore, sch ffs.Scheduler, wm ffs.WalletManager, dc ffs.CidConfig) (*Instance, error) {
	addr, err := wm.NewWallet(ctx, defaultWalletType)
	if err != nil {
		return nil, fmt.Errorf("creating new wallet addr: %s", err)
	}
	config := Config{
		ID:            iid,
		WalletAddr:    addr,
		DefaultConfig: dc,
	}
	ctx, cancel := context.WithCancel(context.Background())
	i := new(iid, confstore, wm, config, sch, ctx, cancel)
	if err := i.store.SaveConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.config.ID, err)
	}
	return i, nil
}

// Load loads a saved Api instance from its ConfigStore.
func Load(iid ffs.InstanceID, confstore ConfigStore, sched ffs.Scheduler, wm ffs.WalletManager) (*Instance, error) {
	config, err := confstore.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("loading instance: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return new(iid, confstore, wm, *config, sched, ctx, cancel), nil
}

func new(iid ffs.InstanceID, confstore ConfigStore, wm ffs.WalletManager, config Config, sch ffs.Scheduler,
	ctx context.Context, cancel context.CancelFunc) *Instance {
	i := &Instance{
		store:   confstore,
		wm:      wm,
		config:  config,
		sched:   sch,
		chSched: sch.Watch(iid),

		cancel:   cancel,
		ctx:      ctx,
		finished: make(chan struct{}),
	}
	go i.watchJobs()
	return i
}

// ID returns the InstanceID of the instance.
func (i *Instance) ID() ffs.InstanceID {
	return i.config.ID
}

// WalletAddr returns the Lotus wallet address of the instance.
func (i *Instance) WalletAddr() string {
	return i.config.WalletAddr
}

func (i *Instance) GetDefaultCidConfig() ffs.CidConfig {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.config.DefaultConfig
}

func (i *Instance) SetDefaultCidConfig(c ffs.CidConfig) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.config.DefaultConfig = c
	return nil
}

// Show returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns ErrNotStore.
func (i *Instance) Show(c cid.Cid) (ffs.CidInfo, error) {
	inf, err := i.store.GetCidInfo(c)
	if err == ErrCidInfoNotFound {
		return inf, ErrNotStored
	}
	if err != nil {
		return inf, fmt.Errorf("getting cid information: %s", err)
	}
	return inf, nil
}

// Info returns instance information
func (i *Instance) Info(ctx context.Context) (InstanceInfo, error) {
	inf := InstanceInfo{
		ID: i.config.ID,
		Wallet: WalletInfo{
			Address: i.config.WalletAddr,
		},
	}
	var err error
	inf.Pins, err = i.store.Cids()
	if err != nil {
		return inf, fmt.Errorf("getting pins from instance: %s", err)
	}

	inf.Wallet.Balance, err = i.wm.Balance(ctx, i.config.WalletAddr)
	if err != nil {
		return inf, fmt.Errorf("getting balance of %s: %s", i.config.WalletAddr, err)
	}

	return inf, nil
}

// Watch subscribes to Job status changes. If jids is empty, it subscribes to
// all Job status changes corresonding to the instance. If jids is not empty,
// it immediately sends current state of those Jobs. If empty, it doesn't.
func (i *Instance) Watch(jids ...ffs.JobID) (<-chan ffs.Job, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	log.Info("registering watcher")

	var jobs []ffs.Job
	for _, jid := range jids {
		j, err := i.sched.GetJob(jid)
		if err != nil {
			return nil, fmt.Errorf("getting current job state: %s", err)
		}
		jobs = append(jobs, j)
	}

	ch := make(chan ffs.Job, 1)
	i.watchers = append(i.watchers, watcher{jobIDs: jids, ch: ch})
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
func (i *Instance) Unwatch(ch <-chan ffs.Job) {
	i.lock.Lock()
	defer i.lock.Unlock()
	for j, w := range i.watchers {
		if w.ch == ch {
			close(w.ch)
			i.watchers[j] = i.watchers[len(i.watchers)-1]
			i.watchers = i.watchers[:len(i.watchers)-1]
			return
		}
	}
}

// AddCid push a new default configuration for the Cid in the Hot and
// Cold layer. (TODO: Soon the configuration will be received as a param,
// to allow different strategies in the Hot and Cold layer. Now a second AddCid
// will error with ErrAlreadyPinned. This might change depending if changing config
// for an existing Cid will use this same API, or another. In any case sounds safer to
// consider some option to specify we want to add a config without overriding some
// existing one.)
func (i *Instance) AddCid(c cid.Cid) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	_, err := i.store.GetCidConfig(c)
	if err == nil {
		return ffs.EmptyJobID, ErrAlreadyPinned
	}
	if err != ErrConfigNotFound {
		return ffs.EmptyJobID, fmt.Errorf("getting cid config: %s", err)
	}

	// TODO: this is a temporal default config for all Cids, this will go
	// away since it will be received as a param.
	cidconf := ffs.AddAction{
		InstanceID: i.config.ID,
		ID:         ffs.NewCidConfigID(),
		Cid:        c,
		Config:     i.config.DefaultConfig,
		Meta: ffs.AddMeta{
			WalletAddr: i.config.WalletAddr,
		},
	}
	log.Infof("adding cid %s to scheduler queue", c)
	jid, err := i.sched.EnqueueCid(cidconf)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.store.PushCidConfig(cidconf); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
}

// Get returns an io.Reader for reading a stored Cid.
// (TODO: this API will change to accept different strategies, like HotOnly,
// UnfreezeCold, or similar. Most prob it will become async, or there will be different
// APIs).
// (TODO: Scheduler.GetFromHot might have to return an error if we want to rate-limit
// hot layer retrievals)
func (i *Instance) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	r, err := i.sched.GetCidFromHot(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting from hot layer %s: %s", c, err)
	}
	return r, nil
}

// Close terminates the running Instance.
func (i *Instance) Close() error {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.cancel()
	<-i.finished
	i.sched.Unwatch(i.chSched)
	return nil
}

func (i *Instance) watchJobs() {
	defer close(i.finished)
	for {
		select {
		case <-i.ctx.Done():
			log.Infof("terminating job watching in %s", i.config.ID)
			return
		case j, ok := <-i.chSched:
			if !ok {
				panic("scheduler closed the watching channel")
			}
			log.Info("received notification from jobstore")
			if err := i.store.SaveCidInfo(j.CidInfo); err != nil {
				log.Errorf("saving cid info %s: %s", j.CidInfo.Cid, err)
				continue
			}
			i.lock.Lock()
			log.Infof("notifying %d subscribed watchers", len(i.watchers))
			for k, w := range i.watchers {
				shouldNotify := len(w.jobIDs) == 0
				for _, jid := range w.jobIDs {
					if jid == j.ID {
						shouldNotify = true
						break
					}
				}
				log.Infof("evaluating watcher %d, shouldNotify %s", k, shouldNotify)
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
