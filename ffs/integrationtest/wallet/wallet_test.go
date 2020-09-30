package wallet

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	"github.com/textileio/powergate/util"
)

const (
	initialBalance int64 = 4000000000000000
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
func TestSendFil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	const amt int64 = 1

	balForAddress := func(addr string) (uint64, error) {
		info, err := fapi.Info(ctx)
		if err != nil {
			return 0, err
		}
		for _, balanceInfo := range info.Balances {
			if balanceInfo.Addr == addr {
				return balanceInfo.Balance, nil
			}
		}
		return 0, fmt.Errorf("no balance info for address %v", addr)
	}

	addrs := fapi.Addrs()
	require.NotEmpty(t, addrs)

	addr1 := addrs[0].Addr

	addr2, err := fapi.NewAddr(ctx, "addr2")
	require.NoError(t, err)

	hasInitialBal := func() bool {
		bal, err := balForAddress(addr2)
		require.NoError(t, err)
		return bal == uint64(initialBalance)
	}

	hasNewBal := func() bool {
		bal, err := balForAddress(addr2)
		require.NoError(t, err)
		return bal == uint64(initialBalance+amt)
	}

	require.Eventually(t, hasInitialBal, time.Second*5, time.Second)

	err = fapi.SendFil(ctx, addr1, addr2, big.NewInt(amt))
	require.NoError(t, err)

	require.Eventually(t, hasNewBal, time.Second*5, time.Second)
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
