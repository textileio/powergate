package manager

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	walletModule "github.com/textileio/powergate/wallet/module"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestNewManager(t *testing.T) {
	t.Parallel()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, tests.NewTxMapDatastore(), addr, false)
	require.NoError(t, err)
	require.NotNil(t, m)
	defer require.NoError(t, cls())
}

func TestNewManagerMasterAddrFFSUseMasterAddr(t *testing.T) {
	t.Parallel()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, tests.NewTxMapDatastore(), addr, true)
	require.NoError(t, err)
	require.NotNil(t, m)
	defer require.NoError(t, cls())
}

func TestNewManagerNoMasterAddrNoFFSUseMasterAddr(t *testing.T) {
	t.Parallel()
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, tests.NewTxMapDatastore(), address.Undef, false)
	require.NoError(t, err)
	require.NotNil(t, m)
	defer require.NoError(t, cls())
}

func TestNewManagerNoMasterAddrFFSUseMasterAddr(t *testing.T) {
	t.Parallel()
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	_, cls, err := newManager(client, tests.NewTxMapDatastore(), address.Undef, true)
	require.Error(t, err)
	defer require.NoError(t, cls())
}

func TestCreate(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, ds, addr, false)
	require.NoError(t, err)
	defer require.NoError(t, cls())

	auth, err := m.Create(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, auth.Token)
	require.True(t, auth.APIID.Valid())
}

func TestCreateWithFFSMasterAddr(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, ds, addr, true)
	require.NoError(t, err)
	defer require.NoError(t, cls())

	auth, err := m.Create(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, auth.Token)
	require.True(t, auth.APIID.Valid())

	i, err := m.GetByAuthToken(auth.Token)
	require.NoError(t, err)
	iAddrs := i.Addrs()
	require.Len(t, iAddrs, 1)
	require.Equal(t, m.wm.MasterAddr().String(), iAddrs[0].Addr)
	c := i.DefaultStorageConfig()
	require.Equal(t, m.wm.MasterAddr().String(), c.Cold.Filecoin.Addr)
}

func TestList(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, ds, addr, false)
	require.NoError(t, err)
	defer require.NoError(t, cls())

	lst, err := m.List()
	require.NoError(t, err)
	require.Equal(t, 0, len(lst))

	auth1, err := m.Create(ctx)
	require.NoError(t, err)
	lst, err = m.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(lst))
	require.Contains(t, lst, auth1)

	auth2, err := m.Create(ctx)
	require.NoError(t, err)

	lst, err = m.List()
	require.NoError(t, err)
	require.Contains(t, lst, auth1)
	require.Contains(t, lst, auth2)
	require.Equal(t, 2, len(lst))
}

func TestGetByAuthToken(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, ds, addr, false)
	require.NoError(t, err)
	defer require.NoError(t, cls())
	auth, err := m.Create(ctx)
	require.NoError(t, err)

	t.Run("Hot", func(t *testing.T) {
		new, err := m.GetByAuthToken(auth.Token)
		require.NoError(t, err)
		require.Equal(t, auth.APIID, new.ID())
		require.NotEmpty(t, new.Addrs())
	})
	t.Run("Cold", func(t *testing.T) {
		client, addr, _ := tests.CreateLocalDevnet(t, 1)
		m, cls, err := newManager(client, ds, addr, false)
		require.NoError(t, err)
		defer require.NoError(t, cls())
		new, err := m.GetByAuthToken(auth.Token)
		require.NoError(t, err)
		require.Equal(t, auth.APIID, new.ID())
		require.NotEmpty(t, new.Addrs())
	})
	t.Run("NonExistant", func(t *testing.T) {
		i, err := m.GetByAuthToken(string("123"))
		require.Equal(t, err, ErrAuthTokenNotFound)
		require.Nil(t, i)
	})
}

func TestDefaultStorageConfig(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	m, cls, err := newManager(client, ds, addr, true)
	require.NoError(t, err)
	defer require.NoError(t, cls())

	// A newly created manager must have
	// the zeroConfig defined value.
	c := m.GetDefaultStorageConfig()
	require.Equal(t, localnetZeroConfig, c)

	// Change the default config and test.
	c.Hot.Enabled = false
	c.Repairable = true
	err = m.SetDefaultStorageConfig(c)
	require.NoError(t, err)
	c2 := m.GetDefaultStorageConfig()
	require.Equal(t, c, c2)
	defer require.NoError(t, cls())

	// Re-open manager, and check that
	// the default config was the last
	// saved one, and not zeroConfig.
	m, cls, err = newManager(client, ds, addr, false)
	require.NoError(t, err)
	defer require.NoError(t, cls())
	c3 := m.GetDefaultStorageConfig()
	require.Equal(t, c, c3)
}

func newManager(clientBuilder lotus.ClientBuilder, ds datastore.TxnDatastore, masterAddr address.Address, ffsUseMasterAddr bool) (*Manager, func() error, error) {
	wm, err := walletModule.New(clientBuilder, masterAddr, *big.NewInt(4000000000), false, "")
	if err != nil {
		return nil, func() error { return nil }, err
	}
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), clientBuilder, util.AvgBlockTime, time.Minute*10)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	m, err := New(ds, wm, dm, nil, ffsUseMasterAddr, true)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	cls := func() error {
		return m.Close()
	}
	return m, cls, nil
}
