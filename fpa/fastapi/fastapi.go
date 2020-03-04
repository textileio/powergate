package fastapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/fpa"
)

var (
	defaultWalletType = "bls"

	log = logging.Logger("fastapi")
)

var (
	ErrAlreadyPinned = errors.New("cid already pinned")
	ErrCantPin       = errors.New("couldn't pin data")
	ErrNotStored     = errors.New("cid not stored")
)

type Instance struct {
	store ConfigStore
	wm    fpa.WalletManager
	hot   fpa.HotLayer

	sched   fpa.Scheduler
	chSched <-chan fpa.Job

	lock     sync.Mutex
	config   Config
	watchers []watcher
	ctx      context.Context
	cancel   context.CancelFunc
	finished chan struct{}
}

type watcher struct {
	jobIDs []fpa.JobID
	ch     chan fpa.Job
}

func New(ctx context.Context, iid fpa.InstanceID, confstore ConfigStore, sch fpa.Scheduler, wm fpa.WalletManager, hot fpa.HotLayer) (*Instance, error) {
	addr, err := wm.NewWallet(ctx, defaultWalletType)
	if err != nil {
		return nil, fmt.Errorf("creating new wallet addr: %s", err)
	}
	config := Config{
		ID:         iid,
		WalletAddr: addr,
	}
	ctx, cancel := context.WithCancel(context.Background())
	i := &Instance{
		store:   confstore,
		wm:      wm,
		config:  config,
		sched:   sch,
		chSched: sch.Watch(iid),
		hot:     hot,

		cancel:   cancel,
		ctx:      ctx,
		finished: make(chan struct{}),
	}
	go i.watchJobs()
	if err := i.store.SaveConfig(config); err != nil {
		return nil, fmt.Errorf("saving new instance %s: %s", i.config.ID, err)
	}
	return i, nil
}

func Load(confstore ConfigStore, sch fpa.Scheduler, wm fpa.WalletManager, hot fpa.HotLayer) (*Instance, error) {
	config, err := confstore.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("loading instance: %s", err)
	}
	return &Instance{
		store:  confstore,
		wm:     wm,
		config: *config,
		sched:  sch,
		hot:    hot,
	}, nil
}

func (i *Instance) ID() fpa.InstanceID {
	return i.config.ID
}

func (i *Instance) WalletAddr() string {
	return i.config.WalletAddr
}

func (i *Instance) Show(c cid.Cid) (fpa.CidInfo, error) {
	inf, exist, err := i.store.GetCidInfo(c)
	if err != nil {
		return inf, fmt.Errorf("getting cid information: %s", err)
	}
	if !exist {
		return inf, ErrNotStored
	}
	return inf, nil
}

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

func (i *Instance) AddFile(ctx context.Context, reader io.Reader) (fpa.JobID, cid.Cid, error) {
	c, err := i.hot.Add(ctx, reader)
	if err != nil {
		return fpa.EmptyJobID, cid.Undef, fmt.Errorf("adding data to hot layer: %s", err)
	}
	jid, err := i.AddCid(c)
	if err != nil {
		return fpa.EmptyJobID, cid.Undef, err
	}
	return jid, c, nil
}

func (i *Instance) Watch(jids ...fpa.JobID) <-chan fpa.Job {
	i.lock.Lock()
	defer i.lock.Unlock()
	log.Info("registering watcher")
	ch := make(chan fpa.Job, 1)
	i.watchers = append(i.watchers, watcher{jobIDs: jids, ch: ch})

	return ch
}

func (i *Instance) Unwatch(ch <-chan fpa.Job) {
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

func (i *Instance) AddCid(c cid.Cid) (fpa.JobID, error) {
	_, err := i.store.GetCidConfig(c)
	if err == nil {
		return fpa.EmptyJobID, ErrAlreadyPinned
	}
	if err != ErrConfigNotFound {
		return fpa.EmptyJobID, fmt.Errorf("getting cid config: %s", err)
	}

	// ToDo: should be a param
	cidconf := fpa.CidConfig{
		ID:         fpa.NewConfigID(),
		InstanceID: i.config.ID,
		Cid:        c,
		Hot: fpa.HotConfig{
			Ipfs: fpa.IpfsConfig{
				Enabled: true,
			},
		},
		Cold: fpa.ColdConfig{
			Filecoin: fpa.FilecoinConfig{
				Enabled:    true,
				WalletAddr: i.config.WalletAddr,
			},
		},
	}
	log.Infof("adding cid %s to scheduler queue", c)
	jid, err := i.sched.Enqueue(cidconf)
	if err != nil {
		return fpa.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.store.PushCidConfig(cidconf); err != nil {
		return fpa.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
}

func (i *Instance) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	r, err := i.hot.Get(ctx, c)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (i *Instance) Close() error {
	i.sched.Unwatch(i.chSched)
	i.cancel()
	<-i.finished
	return nil
}

func (i *Instance) watchJobs() {
	defer close(i.finished)
	for {
		select {
		case <-i.ctx.Done():
			log.Infof("terminating job watching in fastapi %s", i.config.ID)
			return
		case j := <-i.chSched:
			i.lock.Lock()
			log.Info("received notification from jobstore")
			if err := i.store.SaveCidInfo(j.CidInfo); err != nil {
				i.lock.Unlock()
				log.Errorf("saving cid info %s: %s", j.CidInfo.Cid, err)
				continue
			}
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
						log.Warnf("skipping slow fastapi watcher")
					}
				}
			}
			i.lock.Unlock()
		}
	}
}
