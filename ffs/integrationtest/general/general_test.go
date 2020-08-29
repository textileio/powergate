package general

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestAdd(t *testing.T) {
	t.Parallel()
	t.Run("WithDefaultStorageConfig", func(t *testing.T) {
		t.Parallel()
		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ctx := context.Background()
			ipfsAPI, client, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			jid, err := fapi.PushStorageConfig(cid)
			require.NoError(t, err)
			it.RequireJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, nil)
			it.RequireFilStored(ctx, t, client, cid)
			it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
			it.RequireStorageDealRecord(t, fapi, cid)
		})
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		t.Parallel()

		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig().WithHotEnabled(false).WithColdFilDealDuration(util.MinDealDuration + 1234)
			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireStorageDealRecord(t, fapi, cid)
		})
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ipfs, _, fapi, cls := it.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, data := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		rr, err := fapi.Get(ctx, cid)
		require.NoError(t, err)
		fetched, err := ioutil.ReadAll(rr)
		require.NoError(t, err)
		require.True(t, bytes.Equal(data, fetched))
	})
}

func TestInfo(t *testing.T) {
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ipfs, _, fapi, cls := it.NewAPI(t, 1)
		defer cls()

		var err error
		var first api.InstanceInfo
		first, err = fapi.Info(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, first.ID)
		require.Len(t, first.Balances, 1)
		require.NotEmpty(t, first.Balances[0].Addr)
		require.Greater(t, first.Balances[0].Balance, uint64(0))
		require.Equal(t, len(first.Pins), 0)

		r := rand.New(rand.NewSource(22))
		n := 1
		retriesPerDeal := 5
		for i := 0; i < n; i++ {
			for j := 0; j < retriesPerDeal; j++ {
				fmt.Printf("Deal %d, attempt %d...\n", i+1, j+1)
				cid, _ := it.AddRandomFile(t, r, ipfs)
				jid, err := fapi.PushStorageConfig(cid)
				require.NoError(t, err)
				ch := make(chan ffs.Job)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					err = fapi.WatchJobs(ctx, ch, jid)
					close(ch)
				}()
				var failed bool
				stop := false
				for !stop {
					select {
					case <-time.After(120 * time.Second):
						t.Errorf("waiting for job update timeout")
						failed = true
					case job, ok := <-ch:
						require.True(t, ok)
						require.Equal(t, jid, job.ID)
						if job.Status == ffs.Queued || job.Status == ffs.Executing {
							continue
						}
						if job.Status != ffs.Success {
							failed = true
						}
						stop = true
					}
				}
				require.NoError(t, err)
				if !failed {
					it.RequireStorageConfig(t, fapi, cid, nil)
					break
				}
			}
		}

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

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ipfs, _, fapi, cls := it.NewAPI(t, 1)

		defer cls()

		// Test not stored
		c, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")

		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotFound, err)

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		inf, err := fapi.Info(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(inf.Pins))

		c = inf.Pins[0]
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
	tests.RunFlaky(t, func(t *tests.FlakyT) {
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
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
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
	})
}
