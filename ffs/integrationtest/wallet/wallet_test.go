package wallet

import (
	"context"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestAddrs(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	addrs := fapi.Addrs()
	require.Len(t, addrs, 1)
	require.NotEmpty(t, addrs[0].Name)
	require.NotEmpty(t, addrs[0].Addr)
}

func TestNewAddress(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	addr, err := fapi.NewAddr(context.Background(), "my address")
	require.NoError(t, err)
	require.NotEmpty(t, addr)

	addrs := fapi.Addrs()
	require.Len(t, addrs, 2)
}

func TestNewAddressDefault(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	addr, err := fapi.NewAddr(context.Background(), "my address", api.WithMakeDefault(true))
	require.NoError(t, err)
	require.NotEmpty(t, addr)

	defaultConf := fapi.DefaultStorageConfig()
	require.Equal(t, defaultConf.Cold.Filecoin.Addr, addr)
}

func TestSignVerifyMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	addrs := fapi.Addrs()
	bi := addrs[0]

	msg := []byte("hello world")

	sig, err := fapi.SignMessage(ctx, bi.Addr, msg)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	ok, err := fapi.VerifyMessage(ctx, bi.Addr, msg, sig)
	require.NoError(t, err)
	require.True(t, ok)

	newAddr, err := fapi.NewAddr(ctx, "new")
	require.NoError(t, err)
	ok, err = fapi.VerifyMessage(ctx, newAddr, msg, sig)
	require.NoError(t, err)
	require.False(t, ok)
}
