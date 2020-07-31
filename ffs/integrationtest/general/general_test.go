package general

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

const (
	iWalletBal int64 = 4000000000000000
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

func TestGetDefaultStorageConfig(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	defaultConf := fapi.DefaultStorageConfig()
	require.Nil(t, defaultConf.Validate())
}

func TestAdd(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(22))
	t.Run("WithDefaultStorageConfig", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, client, fapi, cls := it.NewAPI(t, 1)
		defer cls()

		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)
		it.RequireFilStored(ctx, t, client, cid)
		it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
		it.RequireStorageDealRecord(t, fapi, cid)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
		defer cls()
		cid, _ := it.AddRandomFile(t, r, ipfsAPI)

		config := fapi.DefaultStorageConfig().WithHotEnabled(false).WithColdFilDealDuration(util.MinDealDuration + 1234)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		it.RequireStorageDealRecord(t, fapi, cid)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, data := it.AddRandomFile(t, r, ipfs)
	jid, err := fapi.PushStorageConfig(cid)
	require.NoError(t, err)
	it.RequireJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, cid, nil)

	t.Run("FromAPI", func(t *testing.T) {
		r, err := fapi.Get(ctx, cid)
		require.NoError(t, err)
		fetched, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		require.True(t, bytes.Equal(data, fetched))
	})
}

func TestInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	var err error
	var first api.InstanceInfo
	t.Run("Minimal", func(t *testing.T) {
		first, err = fapi.Info(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, first.ID)
		require.Len(t, first.Balances, 1)
		require.NotEmpty(t, first.Balances[0].Addr)
		require.Greater(t, first.Balances[0].Balance, uint64(0))
		require.Equal(t, len(first.Pins), 0)
	})

	r := rand.New(rand.NewSource(22))
	n := 3
	for i := 0; i < n; i++ {
		fmt.Println(i)
		cid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)
	}

	t.Run("WithThreeAdd", func(t *testing.T) {
		second, err := fapi.Info(ctx)
		require.NoError(t, err)
		require.Equal(t, second.ID, first.ID)
		require.Len(t, second.Balances, 1)
		require.Equal(t, second.Balances[0].Addr, first.Balances[0].Addr)
		require.Less(t, second.Balances[0].Balance, first.Balances[0].Balance)
		require.Equal(t, n, len(second.Pins))
	})
}

func TestShow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, _, fapi, cls := it.NewAPI(t, 1)

	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotFound, err)
	})

	t.Run("Success", func(t *testing.T) {
		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		inf, err := fapi.Info(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(inf.Pins))

		c := inf.Pins[0]
		s, err := fapi.Show(c)
		require.NoError(t, err)

		require.True(t, s.Cid.Defined())
		require.True(t, time.Now().After(s.Created))
		require.Greater(t, s.Hot.Size, 0)
		require.NotNil(t, s.Hot.Ipfs)
		require.True(t, time.Now().After(s.Hot.Ipfs.Created))
		require.NotNil(t, s.Cold.Filecoin)
		require.True(t, s.Cold.Filecoin.DataCid.Defined())
		require.Equal(t, 1, len(s.Cold.Filecoin.Proposals))
		p := s.Cold.Filecoin.Proposals[0]
		require.True(t, p.ProposalCid.Defined())
		require.Greater(t, p.Duration, int64(0))
		require.Greater(t, p.EpochPrice, uint64(0))
	})
}

func TestColdInstanceLoad(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := it.CreateIPFS(t)
	addr, client, ms := it.NewDevnet(t, 1, ipfsMAddr)
	manager, closeManager := it.NewFFSManager(t, ds, client, addr, ms, ipfs)

	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.

	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)

	ra := rand.New(rand.NewSource(22))
	cid, data := it.AddRandomFile(t, ra, ipfs)
	jid, err := fapi.PushStorageConfig(cid)
	require.NoError(t, err)
	it.RequireJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, cid, nil)

	info, err := fapi.Info(ctx)
	require.NoError(t, err)
	shw, err := fapi.Show(cid)
	require.NoError(t, err)

	// Now close the FFS Instance, and the manager.
	err = fapi.Close()
	require.NoError(t, err)
	closeManager()

	// Rehydrate things again and check state.
	manager, closeManager = it.NewFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()
	fapi, err = manager.GetByAuthToken(auth)
	require.NoError(t, err)

	ninfo, err := fapi.Info(ctx)
	require.NoError(t, err)
	require.Equal(t, info, ninfo)

	nshw, err := fapi.Show(cid)
	require.NoError(t, err)
	require.Equal(t, shw, nshw)

	r, err := fapi.Get(ctx, cid)
	require.NoError(t, err)
	fetched, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	require.True(t, bytes.Equal(data, fetched))
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
		return bal == uint64(iWalletBal)
	}

	hasNewBal := func() bool {
		bal, err := balForAddress(addr2)
		require.NoError(t, err)
		return bal == uint64(iWalletBal+amt)
	}

	require.Eventually(t, hasInitialBal, time.Second*5, time.Second)

	err = fapi.SendFil(ctx, addr1, addr2, big.NewInt(amt))
	require.NoError(t, err)

	require.Eventually(t, hasNewBal, time.Second*5, time.Second)
}

func TestRemove(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c1, _ := it.AddRandomFile(t, r, ipfs)

	config := fapi.DefaultStorageConfig().WithColdEnabled(false)
	jid, err := fapi.PushStorageConfig(c1, api.WithStorageConfig(config))
	require.NoError(t, err)
	it.RequireJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, c1, &config)

	err = fapi.Remove(c1)
	require.Equal(t, api.ErrActiveInStorage, err)

	config = config.WithHotEnabled(false)
	jid, err = fapi.PushStorageConfig(c1, api.WithStorageConfig(config), api.WithOverride(true))
	it.RequireJobState(t, fapi, jid, ffs.Success)
	require.NoError(t, err)

	err = fapi.Remove(c1)
	require.NoError(t, err)
	_, err = fapi.GetStorageConfig(c1)
	require.Equal(t, api.ErrNotFound, err)
}
