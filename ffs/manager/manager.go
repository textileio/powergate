package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/auth"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/util"
)

var (
	// ErrAuthTokenNotFound returns when an auth-token doesn't exist.
	ErrAuthTokenNotFound = errors.New("auth token not found")

	log = logging.Logger("ffs-manager")

	// zeroConfig is a safe-initial value for a default
	// StorageConfig for a manager. A newly (not re-loaded) created
	// manager will have this configuration by default. It can be
	// later changed with the Get/Set APIs. A re-loaded manager will
	// recover its last configured default StorageConfig from the datastore.
	zeroConfig = ffs.StorageConfig{
		Hot: ffs.HotConfig{
			Enabled: true,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 30,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				RepFactor:       5,
				TrustedMiners:   []string{"t016303", "t016304", "t016305", "t016306", "t016309"},
				DealMinDuration: util.MinDealDuration,
			},
		},
	}
	dsDefaultStorageConfigKey = datastore.NewKey("defaultstorageconfig")
)

// Manager creates Api instances, or loads existing ones them from an auth-token.
type Manager struct {
	wm    ffs.WalletManager
	pm    ffs.PaychManager
	drm   ffs.DealRecordsManager
	sched *scheduler.Scheduler

	lock             sync.Mutex
	ds               datastore.Datastore
	auth             *auth.Auth
	instances        map[ffs.APIID]*api.API
	defaultConfig    ffs.StorageConfig
	ffsUseMasterAddr bool

	closed bool
}

// New returns a new Manager.
func New(ds datastore.Datastore, wm ffs.WalletManager, pm ffs.PaychManager, drm ffs.DealRecordsManager, sched *scheduler.Scheduler, ffsUseMasterAddr bool) (*Manager, error) {
	if ffsUseMasterAddr && wm.MasterAddr() == address.Undef {
		return nil, fmt.Errorf("ffsUseMasterAddr requires that master address is defined")
	}
	storageConfig, err := loadDefaultStorageConfig(ds)
	if err != nil {
		return nil, fmt.Errorf("loading default storage config: %s", err)
	}
	return &Manager{
		auth:             auth.New(namespace.Wrap(ds, datastore.NewKey("auth"))),
		ds:               ds,
		wm:               wm,
		pm:               pm,
		drm:              drm,
		sched:            sched,
		instances:        make(map[ffs.APIID]*api.API),
		defaultConfig:    storageConfig,
		ffsUseMasterAddr: ffsUseMasterAddr,
	}, nil
}

// Create creates a new Api instance and an auth-token mapped to it.
func (m *Manager) Create(ctx context.Context) (ffs.APIID, string, error) {
	log.Info("creating instance")

	var addr address.Address
	if m.ffsUseMasterAddr {
		addr = m.wm.MasterAddr()
	} else {
		res, err := m.wm.NewAddress(ctx, "bls")
		if err != nil {
			return ffs.EmptyInstanceID, "", fmt.Errorf("creating new wallet addr: %s", err)
		}
		a, err := address.NewFromString(res)
		if err != nil {
			return ffs.EmptyInstanceID, "", fmt.Errorf("decoding newly created addr: %s", err)
		}
		addr = a
	}

	addrInfo := api.AddrInfo{Name: "Initial Address", Addr: addr.String()}
	switch addr.Protocol() {
	case address.BLS:
		addrInfo.Type = "bls"
	case address.SECP256K1:
		addrInfo.Type = "secp256k1"
	case address.Actor:
		addrInfo.Type = "actor"
	case address.ID:
		addrInfo.Type = "id"
	case address.Unknown:
	default:
		addrInfo.Type = "unknown"
	}

	iid := ffs.NewAPIID()

	fapi, err := api.New(namespace.Wrap(m.ds, datastore.NewKey("api/"+iid.String())), iid, m.sched, m.wm, m.pm, m.drm, m.defaultConfig, addrInfo)
	if err != nil {
		return ffs.EmptyInstanceID, "", fmt.Errorf("creating new instance: %s", err)
	}

	auth, err := m.auth.Generate(fapi.ID())
	if err != nil {
		return ffs.EmptyInstanceID, "", fmt.Errorf("generating auth token for %s: %s", fapi.ID(), err)
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	m.instances[iid] = fapi

	return fapi.ID(), auth, nil
}

// SetDefaultStorageConfig sets the default StorageConfig to be set as default to newly created
// FFS instances.
func (m *Manager) SetDefaultStorageConfig(dc ffs.StorageConfig) error {
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

// GetDefaultStorageConfig returns the current default StorageConfig used
// for newly created FFS instances.
func (m *Manager) GetDefaultStorageConfig() ffs.StorageConfig {
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
func (m *Manager) saveDefaultConfig(dc ffs.StorageConfig) error {
	buf, err := json.Marshal(dc)
	if err != nil {
		return fmt.Errorf("marshaling default config: %s", err)
	}
	if err := m.ds.Put(dsDefaultStorageConfigKey, buf); err != nil {
		return fmt.Errorf("saving default config to datastore: %s", err)
	}
	m.defaultConfig = dc
	return nil
}

func loadDefaultStorageConfig(ds datastore.Datastore) (ffs.StorageConfig, error) {
	d, err := ds.Get(dsDefaultStorageConfigKey)
	if err == datastore.ErrNotFound {
		return zeroConfig, nil
	}
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("get from datastore: %s", err)
	}

	var defaultConfig ffs.StorageConfig
	if err := json.Unmarshal(d, &defaultConfig); err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("unmarshaling default StorageConfig: %s", err)
	}
	return defaultConfig, nil
}
