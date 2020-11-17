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

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
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
			ipfsAPI, client, fapi, cls := itmanager.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			jid, err := fapi.PushStorageConfig(cid)
			require.NoError(t, err)
			it.RequireStorageJobState(t, fapi, jid, ffs.Queued, ffs.Executing)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, nil)
			it.RequireFilStored(ctx, t, client, cid)
			it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
			it.RequireStorageDealRecord(t, fapi, cid)
		})
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		t.Parallel()

		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig().WithHotEnabled(false).WithColdFilDealDuration(util.MinDealDuration + 1234)
			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireStorageJobState(t, fapi, jid, ffs.Queued, ffs.Executing)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireStorageDealRecord(t, fapi, cid)
		})
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, data := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		rr, err := fapi.Get(ctx, cid)
		require.NoError(t, err)
		fetched, err := ioutil.ReadAll(rr)
		require.NoError(t, err)
		require.True(t, bytes.Equal(data, fetched))
	})
}

func TestDealConsistency(t *testing.T) {
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		firstID := fapi.ID()
		require.NotEmpty(t, firstID)

		firstAddrs := fapi.Addrs()
		require.Len(t, firstAddrs, 1)
		require.NotEmpty(t, firstAddrs[0].Addr)

		firstStorageConfigs, err := fapi.GetStorageConfigs()
		require.NoError(t, err)
		require.Equal(t, len(firstStorageConfigs), 0)

		r := rand.New(rand.NewSource(22))
		n := 1
		retriesPerDeal := 5
		for i := 0; i < n; i++ {
			for j := 0; j < retriesPerDeal; j++ {
				fmt.Printf("Deal %d, attempt %d...\n", i+1, j+1)
				cid, _ := it.AddRandomFile(t, r, ipfs)
				jid, err := fapi.PushStorageConfig(cid)
				require.NoError(t, err)
				ch := make(chan ffs.StorageJob)
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

		secondID := fapi.ID()
		require.Equal(t, secondID, firstID)

		secondAddrs := fapi.Addrs()
		require.NoError(t, err)
		require.Len(t, secondAddrs, 1)
		require.Equal(t, secondAddrs[0].Addr, firstAddrs[0].Addr)

		secondStorageConfigs, err := fapi.GetStorageConfigs()
		require.NoError(t, err)
		require.Equal(t, n, len(secondStorageConfigs))
	})
}

func TestShow(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)

		defer cls()

		// Test not stored
		c, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")

		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotFound, err)

		r := rand.New(rand.NewSource(22))
		randomCid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(randomCid)
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, randomCid, nil)

		cfgs, err := fapi.GetStorageConfigs()
		require.NoError(t, err)
		require.Equal(t, 1, len(cfgs))

		cfgCids := make([]cid.Cid, 0, len(cfgs))
		for cid := range cfgs {
			cfgCids = append(cfgCids, cid)
		}

		c = cfgCids[0]

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
		addr, client, ms := itmanager.NewDevnet(t, 1, ipfsMAddr)
		manager, closeManager := itmanager.NewFFSManager(t, ds, client, addr, ms, ipfs)

		auth, err := manager.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3) // Wait for funding txn to finish.

		fapi, err := manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		ra := rand.New(rand.NewSource(22))
		cid, data := it.AddRandomFile(t, ra, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		id := fapi.ID()
		shw, err := fapi.Show(cid)
		require.NoError(t, err)

		// Now close the FFS Instance, and the manager.
		err = fapi.Close()
		require.NoError(t, err)
		closeManager()

		// Rehydrate things again and check state.
		manager, closeManager = itmanager.NewFFSManager(t, ds, client, addr, ms, ipfs)
		defer closeManager()
		fapi, err = manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		nid := fapi.ID()
		require.Equal(t, id, nid)

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

func TestHighMinimumPieceSize(t *testing.T) {
	t.Parallel()
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ds := tests.NewTxMapDatastore()
		ipfs, ipfsMAddr := it.CreateIPFS(t)
		addr, client, ms := itmanager.NewDevnet(t, 1, ipfsMAddr)
		// Set MinimumPieceSize to 1GB so to force failing
		manager, _, closeManager := itmanager.NewCustomFFSManager(t, ds, client, addr, ms, ipfs, 1024*1024*1024)
		defer closeManager()

		auth, err := manager.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3) // Wait for funding txn to finish.

		fapi, err := manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Failed)
	})
}

func TestHotStorageFalseWithAlreadyPinnedData(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ctx := context.Background()
		ds := tests.NewTxMapDatastore()
		ipfs, ipfsMAddr := it.CreateIPFS(t)
		addr, client, ms := itmanager.NewDevnet(t, 1, ipfsMAddr)
		// Set MinimumPieceSize to 1GB so to force failing
		manager, hs, closeManager := itmanager.NewCustomFFSManager(t, ds, client, addr, ms, ipfs, 0)
		defer closeManager()
		auth, err := manager.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3) // Wait for funding txn to finish.

		fapi, err := manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))

		cid, _ := it.AddRandomFileSize(t, r, ipfs, 1600)
		// Simulate staging the data, which pins it.
		_, err = hs.Pin(ctx, fapi.ID(), cid)
		require.NoError(t, err)

		config := fapi.DefaultStorageConfig().WithHotEnabled(false)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		time.Sleep(time.Second)
		it.RequireIpfsUnpinnedCid(ctx, t, cid, ipfs)
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfs, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		c1, _ := it.AddRandomFile(t, r, ipfs)

		config := fapi.DefaultStorageConfig().WithColdEnabled(false)
		jid, err := fapi.PushStorageConfig(c1, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, c1, &config)

		err = fapi.Remove(c1)
		require.Equal(t, api.ErrActiveInStorage, err)

		config = config.WithHotEnabled(false)
		jid, err = fapi.PushStorageConfig(c1, api.WithStorageConfig(config), api.WithOverride(true))
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		require.NoError(t, err)

		err = fapi.Remove(c1)
		require.NoError(t, err)
		_, err = fapi.GetStorageConfigs(c1)
		require.Equal(t, api.ErrNotFound, err)
	})
}

func TestImport(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := itmanager.NewAPI(t, 1)
	defer cls()

	miner := "t01234"
	pieceCid, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
	payloadCid, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC8z")

	// Import correct data, and shouldn't fail.
	imd := []api.ImportDeal{{MinerAddress: miner}}
	err := fapi.ImportStorage(payloadCid, pieceCid, imd)
	require.NoError(t, err)

	// Check that imported data is in fapi2
	i, err := fapi.Show(payloadCid)
	require.NoError(t, err)
	require.False(t, i.Hot.Enabled)
	require.Equal(t, payloadCid, i.Cold.Filecoin.DataCid)
	require.Equal(t, uint64(0), i.Cold.Filecoin.Size)
	require.Len(t, i.Cold.Filecoin.Proposals, 1)

	prop := i.Cold.Filecoin.Proposals[0]
	require.Equal(t, cid.Undef, prop.ProposalCid)
	require.Equal(t, pieceCid, prop.PieceCid)
	require.Equal(t, prop.Duration, int64(0))
	require.Equal(t, prop.ActivationEpoch, int64(0))
	require.Equal(t, prop.StartEpoch, uint64(0))
	require.Equal(t, prop.EpochPrice, uint64(0))
	require.Equal(t, imd[0].MinerAddress, prop.Miner)
}
