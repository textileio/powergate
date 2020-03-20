package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/api/istore"
	"github.com/textileio/powergate/ffs/auth"
)

var (
	// ErrAuthTokenNotFound returns when an auth-token doesn't exist.
	ErrAuthTokenNotFound = errors.New("auth token not found")

	createDefConfig sync.Once
	defCidConfig    ffs.DefaultCidConfig

	istoreNamespace = ds.NewKey("ffs/api/istore")

	log = logging.Logger("ffs-manager")
)

// Manager creates Api instances, or loads existing ones them from an auth-token.
type Manager struct {
	wm    ffs.WalletManager
	sched ffs.Scheduler

	lock      sync.Mutex
	ds        ds.Datastore
	auth      *auth.Auth
	instances map[ffs.ApiID]*api.API

	closed bool
}

// New returns a new Manager.
func New(ds ds.Datastore, wm ffs.WalletManager, sched ffs.Scheduler) (*Manager, error) {
	return &Manager{
		auth:      auth.New(ds),
		ds:        ds,
		wm:        wm,
		sched:     sched,
		instances: make(map[ffs.ApiID]*api.API),
	}, nil
}

// Create creates a new Api instance and an auth-token mapped to it.
func (m *Manager) Create(ctx context.Context) (ffs.ApiID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	createDefConfig.Do(func() {
		defCidConfig = ffs.DefaultCidConfig{
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
					Blacklist:    nil,
				},
			},
		}
	})

	log.Info("creating instance")
	iid := ffs.NewApiID()
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
