package manager

import (
	"context"
	"io"
	"math/big"
	"os"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/tests"
	"github.com/textileio/fil-tools/wallet"
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
		require.NotEmpty(t, new.WalletAddr())
	})
	t.Run("Cold", func(t *testing.T) {
		m, cls := newManager(t, ds)
		defer cls()
		new, err := m.GetByAuthToken(auth)
		require.Nil(t, err)
		require.Equal(t, id, new.ID())
		require.NotEmpty(t, new.WalletAddr())
	})
	t.Run("NonExistant", func(t *testing.T) {
		i, err := m.GetByAuthToken(string("123"))
		require.Equal(t, err, ErrAuthTokenNotFound)
		require.Nil(t, i)
	})
}

func newManager(t *testing.T, ds datastore.TxnDatastore) (*Manager, func()) {
	dnet, addr, _, close := tests.CreateLocalDevnet(t, 1)
	wm, err := wallet.New(dnet.Client, &addr, *big.NewInt(5000000000000))
	require.Nil(t, err)
	m, err := New(ds, wm, &mockSched{})
	require.Nil(t, err)
	cls := func() {
		require.Nil(t, m.Close())
		close()
	}
	return m, cls
}

type mockSched struct{}

func (ms *mockSched) Enqueue(c fpa.CidConfig) (fpa.JobID, error) {
	return fpa.NewJobID(), nil
}
func (ms *mockSched) GetFromHot(ctx context.Context, c cid.Cid) (io.Reader, error) {
	return nil, nil
}
func (ms *mockSched) Watch(iid fpa.InstanceID) <-chan fpa.Job {
	return nil
}
func (ms *mockSched) Unwatch(<-chan fpa.Job) {}
