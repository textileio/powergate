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
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/auth"
	"github.com/textileio/fil-tools/fpa/fastapi"
	"github.com/textileio/fil-tools/fpa/fastapi/store"
)

var (
	ErrAuthTokenNotFound = errors.New("auth token not found")
	bKey                 = ds.NewKey("manager")

	log = logging.Logger("fpa-manager")
)

type Manager struct {
	wm    fpa.WalletManager
	sched fpa.Scheduler
	hot   fpa.HotLayer

	lock      sync.Mutex
	ds        ds.Datastore
	auth      *auth.Repo
	instances map[fpa.InstanceID]*fastapi.Instance

	closed bool
}

func New(ds ds.Datastore, wm fpa.WalletManager, sched fpa.Scheduler, hot fpa.HotLayer) (*Manager, error) {
	return &Manager{
		auth:      auth.New(ds),
		ds:        ds,
		wm:        wm,
		sched:     sched,
		hot:       hot,
		instances: make(map[fpa.InstanceID]*fastapi.Instance),
	}, nil
}

func (m *Manager) Create(ctx context.Context) (fpa.InstanceID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	log.Info("creating instance")

	iid := fpa.NewID()
	confstore := store.New(iid, namespace.Wrap(m.ds, datastore.NewKey("fpa/fastapi/store")))
	fapi, err := fastapi.New(ctx, iid, confstore, m.sched, m.wm, m.hot)
	if err != nil {
		return fpa.EmptyID, "", err
	}
	auth, err := m.auth.Generate(fapi.ID())
	if err != nil {
		return fpa.EmptyID, "", fmt.Errorf("failed generating auth token for %s: %s", fapi.ID(), err)
	}

	m.instances[iid] = fapi

	return fapi.ID(), auth, nil
}

func (m *Manager) GetByAuthToken(token string) (*fastapi.Instance, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	iid, err := m.auth.Get(token)
	if err == auth.ErrNotFound {
		return nil, ErrAuthTokenNotFound
	}
	i, ok := m.instances[iid]
	if !ok {
		log.Infof("loading uncached instance %s", iid)
		confstore := store.New(iid, namespace.Wrap(m.ds, datastore.NewKey("fpa/fastapi/store")))
		i, err = fastapi.Load(confstore, m.sched, m.wm, m.hot)
		if err != nil {
			return nil, fmt.Errorf("loading instance %s: %s", iid, err)
		}
		m.instances[iid] = i
	} else {
		log.Infof("using cached instance %s", iid)
	}

	return i, nil
}

func (m *Manager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		return nil
	}
	return nil
}
