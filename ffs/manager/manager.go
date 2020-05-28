package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/api/istore"
	"github.com/textileio/powergate/ffs/auth"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	// ErrAuthTokenNotFound returns when an auth-token doesn't exist.
	ErrAuthTokenNotFound = errors.New("auth token not found")

	createDefConfig sync.Once
	defCidConfig    ffs.DefaultConfig

	istoreNamespace = datastore.NewKey("ffs/api/istore")

	log = logging.Logger("ffs-manager")
)

// Manager creates Api instances, or loads existing ones them from an auth-token.
type Manager struct {
	wm    ffs.WalletManager
	sched *scheduler.Scheduler

	lock      sync.Mutex
	ds        datastore.Datastore
	auth      *auth.Auth
	instances map[ffs.APIID]*api.API

	closed bool
}

// New returns a new Manager.
func New(ds datastore.Datastore, wm ffs.WalletManager, sched *scheduler.Scheduler) (*Manager, error) {
	return &Manager{
		auth:      auth.New(namespace.Wrap(ds, datastore.NewKey("auth"))),
		ds:        ds,
		wm:        wm,
		sched:     sched,
		instances: make(map[ffs.APIID]*api.API),
	}, nil
}

// Create creates a new Api instance and an auth-token mapped to it.
func (m *Manager) Create(ctx context.Context) (ffs.APIID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	createDefConfig.Do(func() {
		defCidConfig = ffs.DefaultConfig{
			Hot: ffs.HotConfig{
				Enabled: true,
				Ipfs: ffs.IpfsConfig{
					AddTimeout: 30,
				},
			},
			Cold: ffs.ColdConfig{
				Enabled: true,
				Filecoin: ffs.FilConfig{
					RepFactor:    1,
					DealDuration: 1000,
				},
			},
		}
	})

	log.Info("creating instance")
	iid := ffs.NewAPIID()
	is := istore.New(iid, namespace.Wrap(m.ds, istoreNamespace))
	fapi, err := api.New(ctx, iid, is, m.sched, m.wm, defCidConfig)
	if err != nil {
		return ffs.EmptyInstanceID, "", fmt.Errorf("creating new instance: %s", err)
	}

	auth, err := m.auth.Generate(fapi.ID())
	if err != nil {
		return ffs.EmptyInstanceID, "", fmt.Errorf("generating auth token for %s: %s", fapi.ID(), err)
	}

	m.instances[iid] = fapi

	return fapi.ID(), auth, nil
}

// List returns a list of all existing API instances.
func (m *Manager) List() ([]ffs.APIID, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	res, err := m.auth.List()
	if err != nil {
		return nil, fmt.Errorf("listing existing instances: %s", err)
	}
	return res, nil
}

// GetByAuthToken loads an existing instance using an auth-token. If auth-token doesn't exist,
// it returns ErrAuthTokenNotFound.
func (m *Manager) GetByAuthToken(token string) (*api.API, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	iid, err := m.auth.Get(token)
	if err == auth.ErrNotFound {
		return nil, ErrAuthTokenNotFound
	}

	i, ok := m.instances[iid]
	if !ok {
		log.Infof("loading uncached instance %s", iid)
		is := istore.New(iid, namespace.Wrap(m.ds, istoreNamespace))
		i, err = api.Load(iid, is, m.sched, m.wm)
		if err != nil {
			return nil, fmt.Errorf("loading instance %s: %s", iid, err)
		}
		m.instances[iid] = i
	} else {
		log.Infof("using cached instance %s", iid)
	}

	return i, nil
}

// Close closes a Manager and consequently all loaded instances.
func (m *Manager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		return nil
	}
	for _, i := range m.instances {
		if err := i.Close(); err != nil {
			log.Errorf("closing instance %s: %s", i.ID(), err)
		}
	}
	m.closed = true
	return nil
}
