package config

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestSetDefaultStorageConfig(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	config := ffs.StorageConfig{
		Hot: ffs.HotConfig{
			Enabled: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 10,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				DealMinDuration: util.MinDealDuration + 22333,
				RepFactor:       23,
				Addr:            "123456",
			},
		},
	}
	err := fapi.SetDefaultStorageConfig(config)
	require.NoError(t, err)
	newConfig := fapi.DefaultStorageConfig()
	require.Equal(t, newConfig.Hot, config.Hot)
	require.Equal(t, newConfig.Cold, config.Cold)
}

func TestEnabledConfigChange(t *testing.T) {
	t.Parallel()
	t.Run("HotEnabledDisabled", func(t *testing.T) {
		t.Parallel()
		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ctx := context.Background()
			ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig()

			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)

			config = fapi.DefaultStorageConfig().WithHotEnabled(false)
			jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)
		})
	})
	t.Run("HotDisabledEnabled", func(t *testing.T) {
		t.Parallel()
		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ctx := context.Background()
			ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig().WithHotEnabled(false)

			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)

			config = fapi.DefaultStorageConfig().WithHotEnabled(true)
			jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
		})
	})
	t.Run("ColdDisabledEnabled", func(t *testing.T) {
		t.Parallel()

		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ctx := context.Background()
			ipfsAPI, client, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig().WithColdEnabled(false)

			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireFilUnstored(ctx, t, client, cid)

			config = fapi.DefaultStorageConfig().WithHotEnabled(true)
			jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireFilStored(ctx, t, client, cid)
		})
	})
	t.Run("ColdEnabledDisabled", func(t *testing.T) {
		t.Parallel()

		tests.RunFlaky(t, func(t *tests.FlakyT) {
			ctx := context.Background()
			ipfsAPI, client, fapi, cls := it.NewAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := it.AddRandomFile(t, r, ipfsAPI)
			config := fapi.DefaultStorageConfig().WithColdEnabled(false)

			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)
			it.RequireStorageConfig(t, fapi, cid, &config)
			it.RequireFilUnstored(ctx, t, client, cid)

			config = fapi.DefaultStorageConfig().WithHotEnabled(true)
			jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Success)

			// Yes, still stored in filecoin since deals can't be
			// undone.
			it.RequireFilStored(ctx, t, client, cid)
			// Despite of the above, check that the Cid Config still reflects
			// that this *shouldn't* be in cold storage. To indicate
			// this can't be renewed, or any other future action that tries to
			// store it again in cold storage.
			it.RequireStorageConfig(t, fapi, cid, &config)
		})
	})
}

func TestFilecoinEnableConfig(t *testing.T) {
	t.Parallel()
	tableTest := []struct {
		HotEnabled  bool
		ColdEnabled bool
	}{
		{HotEnabled: true, ColdEnabled: true},
		{HotEnabled: false, ColdEnabled: true},
		{HotEnabled: true, ColdEnabled: false},
		{HotEnabled: false, ColdEnabled: false},
	}

	for _, tt := range tableTest {
		tt := tt
		name := fmt.Sprintf("Hot(%v)/Cold(%v)", tt.HotEnabled, tt.ColdEnabled)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tests.RunFlaky(t, func(t *tests.FlakyT) {
				ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
				defer cls()

				r := rand.New(rand.NewSource(22))
				cid, _ := it.AddRandomFile(t, r, ipfsAPI)
				config := fapi.DefaultStorageConfig().WithColdEnabled(tt.ColdEnabled).WithHotEnabled(tt.HotEnabled)

				jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
				require.NoError(t, err)

				expectedJobState := ffs.Success
				it.RequireEventualJobState(t, fapi, jid, expectedJobState)

				if expectedJobState == ffs.Success {
					it.RequireStorageConfig(t, fapi, cid, &config)

					// Show() assertions
					cinfo, err := fapi.Show(cid)
					require.NoError(t, err)
					require.Equal(t, tt.HotEnabled, cinfo.Hot.Enabled)
					if tt.ColdEnabled {
						require.NotEmpty(t, cinfo.Cold.Filecoin.Proposals)
					} else {
						require.Empty(t, cinfo.Cold.Filecoin.Proposals)
					}

					// Get() assertions
					ctx := context.Background()
					_, err = fapi.Get(ctx, cid)
					var expectedErr error
					if !tt.HotEnabled {
						expectedErr = api.ErrHotStorageDisabled
					}
					require.Equal(t, expectedErr, err)

					// External assertions
					if !tt.HotEnabled {
						it.RequireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)
					} else {
						it.RequireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
					}
				}
			})
		})
	}
}

func TestHotTimeoutConfig(t *testing.T) {
	scheduler.HardcodedHotTimeout = time.Second * 10
	t.SkipNow()
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	t.Run("ShortTime", func(t *testing.T) {
		tests.RunFlaky(t, func(t *tests.FlakyT) {
			cid, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
			config := fapi.DefaultStorageConfig().WithHotIpfsAddTimeout(1)
			jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
			require.NoError(t, err)
			it.RequireEventualJobState(t, fapi, jid, ffs.Failed)
		})
	})
}

func TestDurationConfig(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfsAPI, _, fapi, cls := it.NewAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		duration := int64(util.MinDealDuration + 1234)
		config := fapi.DefaultStorageConfig().WithColdFilDealDuration(duration)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		p := cinfo.Cold.Filecoin.Proposals[0]
		require.Greater(t, p.Duration, duration)
		require.Greater(t, p.ActivationEpoch, int64(0))
	})
}

func TestGetDefaultStorageConfig(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	defaultConf := fapi.DefaultStorageConfig()
	require.Nil(t, defaultConf.Validate())
}
