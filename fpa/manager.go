package fpa

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/textileio/fil-tools/deals"
	"github.com/textileio/fil-tools/fpa/auth"
	fa "github.com/textileio/fil-tools/fpa/fastapi"
	"github.com/textileio/fil-tools/fpa/types"
)

var (
	ErrAuthTokenNotFound = errors.New("auth token not found")
	bKey                 = ds.NewKey("manager")

	log = logging.Logger("fpa-manager")
)

type Manager struct {
	wm   types.WalletManager
	dm   *deals.Module
	ms   types.MinerSelector
	au   types.Auditer
	ipfs iface.CoreAPI

	lock sync.Mutex
	ds   ds.Datastore
	auth *auth.Repo

	closed bool
}

func New(ds ds.Datastore, wm types.WalletManager, dm *deals.Module, ms types.MinerSelector, au types.Auditer, ipfs iface.CoreAPI) (*Manager, error) {
	return &Manager{
		auth: auth.New(ds),
		ds:   ds,
		ipfs: ipfs,
		wm:   wm,
		dm:   dm,
		ms:   ms,
		au:   au,
	}, nil
}

func (m *Manager) Create(ctx context.Context) (fa.ID, string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	log.Info("creating instance")

	fapi, err := fa.New(ctx, m.ds, m.ipfs, m.dm, m.ms, m.au, m.wm)
	if err != nil {
		return fa.EmptyID, "", err
	}
	auth, err := m.auth.Generate(fapi.ID())
	if err != nil {
		return fa.EmptyID, "", fmt.Errorf("failed generating auth token for %s: %s", fapi.ID(), err)
	}

	return fapi.ID(), auth, nil
}

func (m *Manager) GetByAuthToken(token string) (*fa.Instance, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	id, err := m.auth.Get(token)
	if err == auth.ErrNotFound {
		return nil, ErrAuthTokenNotFound
	}
	return fa.LoadFromID(m.ds, m.ipfs, m.dm, m.ms, m.au, m.wm, id)
}

func (m *Manager) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		return nil
	}
	return nil
}
