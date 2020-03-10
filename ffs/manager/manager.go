package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/api/store"
	"github.com/textileio/powergate/ffs/auth"
)

var (
	// ErrAuthTokenNotFound returns when an auth-token doesn't exist.
	ErrAuthTokenNotFound = errors.New("auth token not found")

	createDefConfig   sync.Once
	defInstanceConfig ffs.CidConfig

	log = logging.Logger("ffs-manager")
)

// Manager creates Api instances, or loads existing ones them from an auth-token.
type Manager struct {
	wm    ffs.WalletManager
	sched ffs.Scheduler

	lock      sync.Mutex
	ds        ds.Datastore
	auth      *auth.Auth
	instances map[ffs.InstanceID]*api.Instance

	closed bool
}

// New returns a new Manager.
func New(ds ds.Datastore, wm ffs.WalletManager, sched ffs.Scheduler) (*Manager, error) {
	return &Manager{
		auth:      auth.New(ds),
		ds:        ds,
		wm:        wm,
		sched:     sched,
		instances: make(map[ffs.InstanceID]*api.Instance),
	}, nil
}

// Create creates a new Api instance and an auth-token mapped to it.
func (m *Manager) Create(ctx context.Context) (ffs.InstanceID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	createDefConfig.Do(func() {
		defIpfsConfig, err := ffs.NewIpfsConfig(true)
		if err != nil {
			log.Fatalf("creating default ipfs config: %s", err)
		}
		defFilecoinConfig, err := ffs.NewFilecoinConfig(true, 1)
		if err != nil {
			log.Fatalf("creating default filecoin config: %s", err)
		}
		defInstanceConfig = ffs.CidConfig{
			Hot: ffs.HotConfig{
				Ipfs: defIpfsConfig,
			},
			Cold: ffs.ColdConfig{
				Filecoin: defFilecoinConfig,
			},
		}
	})

	log.Info("creating instance")
	iid := ffs.NewInstanceID()
	confstore := store.New(iid, namespace.Wrap(m.ds, datastore.NewKey("ffs/api/store")))
	fapi, err := api.New(ctx, iid, confstore, m.sched, m.wm, defInstanceConfig)
	if err != nil {
		return ffs.EmptyID, "", fmt.Errorf("creating new instance: %s", err)
	}

	auth, err := m.auth.Generate(fapi.ID())
	if err != nil {
		return ffs.EmptyID, "", fmt.Errorf("generating auth token for %s: %s", fapi.ID(), err)
	}

	m.instances[iid] = fapi

	return fapi.ID(), auth, nil
}

// GetByAuthToken loads an existing instance using an auth-token. If auth-token doesn't exist,
// it returns ErrAuthTokenNotFound.
func (m *Manager) GetByAuthToken(token string) (*api.Instance, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	iid, err := m.auth.Get(token)
	if err == auth.ErrNotFound {
		return nil, ErrAuthTokenNotFound
	}

	i, ok := m.instances[iid]
	if !ok {
		log.Infof("loading uncached instance %s", iid)
		confstore := store.New(iid, namespace.Wrap(m.ds, datastore.NewKey("ffs/api/store")))
		i, err = api.Load(iid, confstore, m.sched, m.wm)
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
