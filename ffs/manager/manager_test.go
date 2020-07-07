package manager

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	dealsModule "github.com/textileio/powergate/deals/module"
	paych "github.com/textileio/powergate/paych/lotus"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	walletModule "github.com/textileio/powergate/wallet/module"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestNewManager(t *testing.T) {
	t.Parallel()
	m, cls := newManager(t, tests.NewTxMapDatastore())
	require.NotNil(t, m)
	defer cls()
}

func TestCreate(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	m, cls := newManager(t, ds)
	defer cls()

	id, auth, err := m.Create(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, auth)
	require.True(t, id.Valid())
}

func TestList(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	m, cls := newManager(t, ds)
	defer cls()

	lst, err := m.List()
	require.Nil(t, err)
	require.Equal(t, 0, len(lst))

	id1, _, err := m.Create(ctx)
	require.Nil(t, err)
	lst, err = m.List()
	require.Nil(t, err)
	require.Equal(t, 1, len(lst))
	require.Contains(t, lst, id1)

	id2, _, err := m.Create(ctx)
	require.Nil(t, err)

	lst, err = m.List()
	require.Nil(t, err)
	require.Contains(t, lst, id1)
	require.Contains(t, lst, id2)
	require.Equal(t, 2, len(lst))
}

func TestGetByAuthToken(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ctx := context.Background()
	m, cls := newManager(t, ds)
	defer cls()
	id, auth, err := m.Create(ctx)
	require.Nil(t, err)

	t.Run("Hot", func(t *testing.T) {
		new, err := m.GetByAuthToken(auth)
		require.Nil(t, err)
		require.Equal(t, id, new.ID())
		require.NotEmpty(t, new.Addrs())
	})
	t.Run("Cold", func(t *testing.T) {
		m, cls := newManager(t, ds)
		defer cls()
		new, err := m.GetByAuthToken(auth)
		require.Nil(t, err)
		require.Equal(t, id, new.ID())
		require.NotEmpty(t, new.Addrs())
	})
	t.Run("NonExistant", func(t *testing.T) {
		i, err := m.GetByAuthToken(string("123"))
		require.Equal(t, err, ErrAuthTokenNotFound)
		require.Nil(t, i)
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	m, cls := newManager(t, ds)
	defer cls()

	// A newly created manager must have
	// the zeroConfig defined value.
	c := m.GetDefaultConfig()
	require.Equal(t, zeroConfig, c)

	// Change the default config and test.
	c.Hot.Enabled = false
	c.Repairable = true
	err := m.SetDefaultConfig(c)
	require.NoError(t, err)
	c2 := m.GetDefaultConfig()
	require.Equal(t, c, c2)
	cls()

	// Re-open manager, and check that
	// the default config was the last
	// saved one, and not zeroConfig.
	m, cls = newManager(t, ds)
	defer cls()
	c3 := m.GetDefaultConfig()
	require.Equal(t, c, c3)
}

func newManager(t *testing.T, ds datastore.TxnDatastore) (*Manager, func()) {
	client, addr, _ := tests.CreateLocalDevnet(t, 1)
	wm, err := walletModule.New(client, addr, *big.NewInt(4000000000), false)
	require.Nil(t, err)
	pm := paych.New(client)
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), client)
	require.Nil(t, err)
	m, err := New(ds, wm, pm, dm, nil)
	require.Nil(t, err)
	cls := func() {
		require.Nil(t, m.Close())
	}
	return m, cls
}
