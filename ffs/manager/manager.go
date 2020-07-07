package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/auth"
	"github.com/textileio/powergate/ffs/scheduler"
)

var (
	// ErrAuthTokenNotFound returns when an auth-token doesn't exist.
	ErrAuthTokenNotFound = errors.New("auth token not found")

	log = logging.Logger("ffs-manager")

	// zeroConfig is a safe-initial value for a default
	// CidConfig for a manager. A newly (not re-loaded) created
	// manager will have this configuration by default. It can be
	// later changed with the Get/Set APIs. A re-loaded manager will
	// recover its laste configured CidConfig from the datastore.
	zeroConfig = ffs.DefaultConfig{
		Hot: ffs.HotConfig{
			Enabled: true,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 30,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				RepFactor:       1,
				DealMinDuration: 1000,
			},
		},
	}
	dsDefaultCidConfigKey = datastore.NewKey("defaultcidconfig")
)

// Manager creates Api instances, or loads existing ones them from an auth-token.
type Manager struct {
	wm    ffs.WalletManager
	pm    ffs.PaychManager
	drm   ffs.DealRecordsManager
	sched *scheduler.Scheduler

	lock          sync.Mutex
	ds            datastore.Datastore
	auth          *auth.Auth
	instances     map[ffs.APIID]*api.API
	defaultConfig ffs.DefaultConfig

	closed bool
}

// New returns a new Manager.
func New(ds datastore.Datastore, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager, sched *scheduler.Scheduler) (*Manager, error) {
	cidConfig, err := loadDefaultCidConfig(ds)
	if err != nil {
		return nil, fmt.Errorf("loading default cidconfig: %s", err)
	}
	return &Manager{
		auth:          auth.New(namespace.Wrap(ds, datastore.NewKey("auth"))),
		ds:            ds,
		wm:            wm,
		pm:            pm,
		drm:           drm,
		sched:         sched,
		instances:     make(map[ffs.APIID]*api.API),
		defaultConfig: cidConfig,
	}, nil
}

// Create creates a new Api instance and an auth-token mapped to it.
func (m *Manager) Create(ctx context.Context) (ffs.APIID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	log.Info("creating instance")
	iid := ffs.NewAPIID()
	fapi, err := api.New(ctx, namespace.Wrap(m.ds, datastore.NewKey("api/"+iid.String())), iid, m.sched, m.wm, m.pm, m.drm, m.defaultConfig)
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

// SetDefaultConfig sets the default CidConfig to be set as default to newly created
// FFS instances.
func (m *Manager) SetDefaultConfig(dc ffs.DefaultConfig) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if err := m.saveDefaultConfig(dc); err != nil {
		return fmt.Errorf("persisting default configuration: %s", err)
	}
	return nil
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
		i, err = api.Load(namespace.Wrap(m.ds, datastore.NewKey("api/"+iid.String())), iid, m.sched, m.wm, m.pm, m.drm)
		if err != nil {
			return nil, fmt.Errorf("loading instance %s: %s", iid, err)
		}
		m.instances[iid] = i
	} else {
		log.Infof("using cached instance %s", iid)
	}

	return i, nil
}

// GetDefaultConfig returns the current default CidConfig used
// for newly created FFS instances.
func (m *Manager) GetDefaultConfig() ffs.DefaultConfig {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.defaultConfig
}

// Close closes a Manager and consequently all loaded instances.
func (m *Manager) Close() error {
	log.Info("closing...")
	defer log.Info("closed")
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

// saveDefaultConfig persists a new default configuration and updates
// the cached value. This method must be guarded.
func (m *Manager) saveDefaultConfig(dc ffs.DefaultConfig) error {
	buf, err := json.Marshal(dc)
	if err != nil {
		return fmt.Errorf("marshaling default config: %s", err)
	}
	if err := m.ds.Put(dsDefaultCidConfigKey, buf); err != nil {
		return fmt.Errorf("saving default config to datastore: %s", err)
	}
	m.defaultConfig = dc
	return nil
}

func loadDefaultCidConfig(ds datastore.Datastore) (ffs.DefaultConfig, error) {
	d, err := ds.Get(dsDefaultCidConfigKey)
	if err == datastore.ErrNotFound {
		return zeroConfig, nil
	}
	if err != nil {
		return ffs.DefaultConfig{}, fmt.Errorf("get from datastore: %s", err)
	}

	var defaultConfig ffs.DefaultConfig
	if err := json.Unmarshal(d, &defaultConfig); err != nil {
		return ffs.DefaultConfig{}, fmt.Errorf("unmarshaling default cidconfig: %s", err)
	}
	return defaultConfig, nil
}
