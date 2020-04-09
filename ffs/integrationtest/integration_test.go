package integrationtest

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

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
	"github.com/textileio/lotus-client/api/apistruct"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/api/istore"
	"github.com/textileio/powergate/ffs/cidlogger"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/filcold/lotuschain"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/ffs/scheduler/cistore"
	"github.com/textileio/powergate/ffs/scheduler/jstore"
	"github.com/textileio/powergate/ffs/scheduler/pcstore"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	"github.com/textileio/powergate/wallet"
)

const (
	tmpDir = "/tmp/powergate"
)

func TestMain(m *testing.M) {
	_ = os.RemoveAll(tmpDir)
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		os.Mkdir(tmpDir, os.ModePerm)
	}

	logging.SetAllLoggers(logging.LevelWarn)
	// logging.SetLogLevel("scheduler", "debug")
	// logging.SetLogLevel("api", "debug")
	// logging.SetLogLevel("jobstore", "debug")
	// logging.SetLogLevel("coreipfs", "debug")

	os.Exit(m.Run())
}

func TestSetDefaultConfig(t *testing.T) {
	_, fapi, cls := newAPI(t, 1)
	defer cls()

	config := ffs.DefaultCidConfig{
		Hot: ffs.HotConfig{
			Enabled: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 10,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				DealDuration: 22333,
				RepFactor:    23,
			},
		},
	}
	err := fapi.SetDefaultCidConfig(config)
	require.Nil(t, err)
	newConfig := fapi.GetDefaultCidConfig(cid.Undef)
	require.Equal(t, newConfig.Hot, config.Hot)
	require.Equal(t, newConfig.Cold, config.Cold)
}

func TestAdd(t *testing.T) {
	r := rand.New(rand.NewSource(22))
	t.Run("WithDefaultConfig", func(t *testing.T) {
		ctx := context.Background()
		ipfsDocker, cls := tests.LaunchDocker()
		defer cls()
		ds := tests.NewTxMapDatastore()
		addr, client, ms := newDevnet(t, 1)
		ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
		defer closeInternal()

		cid, _ := addRandomFile(t, r, ipfsAPI)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)
		requireFilStored(t, ctx, client, cid)
		requireIpfsPinnedCid(t, ctx, cid, ipfsAPI)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		ipfsAPI, fapi, cls := newAPI(t, 1)
		defer cls()
		cid, _ := addRandomFile(t, r, ipfsAPI)

		config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false).WithColdFilDealDuration(int64(1234))
		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
	})
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, r, ipfs)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, nil)

	t.Run("FromAPI", func(t *testing.T) {
		r, err := fapi.Get(ctx, cid)
		require.Nil(t, err)
		fetched, err := ioutil.ReadAll(r)
		require.Nil(t, err)
		require.True(t, bytes.Equal(data, fetched))
	})
}

func TestInfo(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newAPI(t, 1)
	defer cls()

	var err error
	var first api.InstanceInfo
	t.Run("Minimal", func(t *testing.T) {
		first, err = fapi.Info(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, first.ID)
		require.NotEmpty(t, first.Wallet.Address)
		require.Greater(t, first.Wallet.Balance, uint64(0))
		require.Equal(t, len(first.Pins), 0)
	})

	r := rand.New(rand.NewSource(22))
	n := 1
	for i := 0; i < n; i++ {
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)
	}

	t.Run("WithOneAdd", func(t *testing.T) {
		second, err := fapi.Info(ctx)
		require.Nil(t, err)
		require.Equal(t, second.ID, first.ID)
		require.Equal(t, second.Wallet.Address, first.Wallet.Address)
		require.Less(t, second.Wallet.Balance, first.Wallet.Balance)
		require.Equal(t, n, len(second.Pins))
	})
}

func TestShow(t *testing.T) {
	ctx := context.Background()
	ipfs, fapi, cls := newAPI(t, 1)

	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotStored, err)
	})

	t.Run("Success", func(t *testing.T) {
		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)

		inf, err := fapi.Info(ctx)
		require.Nil(t, err)
		require.Equal(t, 1, len(inf.Pins))

		c := inf.Pins[0]
		s, err := fapi.Show(c)
		require.Nil(t, err)

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
	})
}

func TestColdInstanceLoad(t *testing.T) {
	ctx := context.Background()
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })

	ds := tests.NewTxMapDatastore()
	addr, client, ms := newDevnet(t, 1)

	ipfsAPI, fapi, cls := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	ra := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, ra, ipfsAPI)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, nil)

	info, err := fapi.Info(ctx)
	require.Nil(t, err)
	shw, err := fapi.Show(cid)
	require.Nil(t, err)
	cls()

	_, fapi, cls = newAPIFromDs(t, ds, fapi.ID(), client, addr, ms, ipfsDocker)
	defer cls()
	ninfo, err := fapi.Info(ctx)
	require.Nil(t, err)
	require.Equal(t, info, ninfo)

	nshw, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, shw, nshw)

	r, err := fapi.Get(ctx, cid)
	require.Nil(t, err)
	fetched, err := ioutil.ReadAll(r)
	require.Nil(t, err)
	require.True(t, bytes.Equal(data, fetched))
}

func TestRepFactor(t *testing.T) {
	rfs := []int{1, 2}
	r := rand.New(rand.NewSource(22))
	for _, rf := range rfs {
		t.Run(fmt.Sprintf("%d", rf), func(t *testing.T) {
			ipfsAPI, fapi, cls := newAPI(t, 2)
			defer cls()
			cid, _ := addRandomFile(t, r, ipfsAPI)
			config := fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(rf)
			jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
			require.Nil(t, err)
			requireJobState(t, fapi, jid, ffs.Success)
			requireCidConfig(t, fapi, cid, &config)

			cinfo, err := fapi.Show(cid)
			require.Nil(t, err)
			require.Equal(t, rf, len(cinfo.Cold.Filecoin.Proposals))
		})
	}
}

func TestRepFactorIncrease(t *testing.T) {
	// ToDo: unskip when testnet/3  allows more than one deal
	// See https://bit.ly/2JxQSQk
	t.SkipNow()

	r := rand.New(rand.NewSource(22))
	ipfsAPI, fapi, cls := newAPI(t, 2)
	defer cls()
	cid, _ := addRandomFile(t, r, ipfsAPI)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, nil)

	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 1, len(cinfo.Cold.Filecoin.Proposals))
	firstProposal := cinfo.Cold.Filecoin.Proposals[0]

	config := fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(2)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err = fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))
	require.Contains(t, cinfo.Cold.Filecoin.Proposals, firstProposal)
}

func TestRepFactorDecrease(t *testing.T) {
	r := rand.New(rand.NewSource(22))
	ipfsAPI, fapi, cls := newAPI(t, 2)
	defer cls()

	cid, _ := addRandomFile(t, r, ipfsAPI)
	config := fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(2)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))

	config = fapi.GetDefaultCidConfig(cid).WithColdFilRepFactor(1)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	cinfo, err = fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))
}

func TestHotTimeoutConfig(t *testing.T) {
	_, fapi, cls := newAPI(t, 1)
	defer cls()

	t.Run("ShortTime", func(t *testing.T) {
		cid, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		config := fapi.GetDefaultCidConfig(cid).WithHotIpfsAddTimeout(1)
		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Failed)
	})
}

func TestDurationConfig(t *testing.T) {
	ipfsAPI, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsAPI)
	duration := int64(1234)
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(duration)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, duration, p.Duration)
	require.Greater(t, p.ActivationEpoch, int64(0))
}

func TestFilecoinExcludedMiners(t *testing.T) {
	ipfsAPI, fapi, cls := newAPI(t, 2)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsAPI)
	excludedMiner := "t01000"
	config := fapi.GetDefaultCidConfig(cid).WithColdFilExcludedMiners([]string{excludedMiner})

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.NotEqual(t, p.Miner, excludedMiner)
}

func TestFilecoinCountryFilter(t *testing.T) {
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })

	countries := []string{"China", "Uruguay"}
	numMiners := len(countries)
	client, addr, _ := tests.CreateLocalDevnet(t, numMiners)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 1000+i)
	}
	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: countries[i], EpochPrice: 1000000}
	}
	ms := fixed.New(fixedMiners)
	ds := tests.NewTxMapDatastore()
	ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	defer closeInternal()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsAPI)
	countryFilter := []string{"Uruguay"}
	config := fapi.GetDefaultCidConfig(cid).WithColdFilCountryCodes(countryFilter)

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, p.Miner, "t01001")
}

func TestFilecoinEnableConfig(t *testing.T) {
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
		name := fmt.Sprintf("Hot(%v)/Cold(%v)", tt.HotEnabled, tt.ColdEnabled)
		t.Run(name, func(t *testing.T) {
			ipfsAPI, fapi, cls := newAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := addRandomFile(t, r, ipfsAPI)
			config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(tt.ColdEnabled).WithHotEnabled(tt.HotEnabled)

			var expectedErr error
			if !tt.HotEnabled && !tt.ColdEnabled {
				expectedErr = ffs.ErrBothStoragesDisabled
			}
			jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
			require.Equal(t, expectedErr, err)
			if expectedErr != nil {
				return
			}

			expectedJobState := ffs.Success
			var errCause string
			if !tt.ColdEnabled && !tt.HotEnabled {
				expectedJobState = ffs.Failed
				errCause = ffs.ErrBothStoragesDisabled.Error()
			}
			job := requireJobState(t, fapi, jid, expectedJobState)
			require.Equal(t, errCause, job.ErrCause)

			if expectedJobState == ffs.Success {
				requireCidConfig(t, fapi, cid, &config)

				// Show() assertions
				cinfo, err := fapi.Show(cid)
				require.Nil(t, err)
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
					expectedErr = ffs.ErrHotStorageDisabled
				}
				require.Equal(t, expectedErr, err)

				// External assertions
				if !tt.HotEnabled {
					requireIpfsUnpinnedCid(t, ctx, cid, ipfsAPI)
				} else {
					requireIpfsPinnedCid(t, ctx, cid, ipfsAPI)
				}
			}

		})
	}
}

func requireIpfsUnpinnedCid(t *testing.T, ctx context.Context, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)
	for _, p := range pins {
		require.NotEqual(t, cid, p.Path().Cid(), "Cid isn't unpined from IPFS node")
	}
}

func requireIpfsPinnedCid(t *testing.T, ctx context.Context, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)

	pinned := false
	for _, p := range pins {
		if p.Path().Cid() == cid {
			pinned = true
			break
		}
	}
	require.True(t, pinned, "Cid should be pinned in IPFS node")
}

func TestEnabledConfigChange(t *testing.T) {
	t.Run("HotEnabledDisabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsPinnedCid(t, ctx, cid, ipfsAPI)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(false)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsUnpinnedCid(t, ctx, cid, ipfsAPI)
	})
	t.Run("HotDisabledEnabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsUnpinnedCid(t, ctx, cid, ipfsAPI)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsPinnedCid(t, ctx, cid, ipfsAPI)
	})
	t.Run("ColdDisabledEnabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsDocker, cls := tests.LaunchDocker()
		t.Cleanup(func() { cls() })
		ds := tests.NewTxMapDatastore()
		addr, client, ms := newDevnet(t, 1)
		ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
		t.Cleanup(func() { closeInternal() })

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilUnstored(t, ctx, client, cid)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilStored(t, ctx, client, cid)

	})
	t.Run("ColdEnabledDisabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsDocker, cls := tests.LaunchDocker()
		t.Cleanup(func() { cls() })
		ds := tests.NewTxMapDatastore()
		addr, client, ms := newDevnet(t, 1)
		ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
		t.Cleanup(func() { closeInternal() })

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilUnstored(t, ctx, client, cid)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)

		// Yes, still stored in filecoin since deals can't be
		// undone.
		requireFilStored(t, ctx, client, cid)
		// Despite of the above, check that the Cid Config still reflects
		// that this *shouldn't* be in the Cold Storage. To indicate
		// this can't be renewed, or any other future action that tries to
		// store it again in the Cold Layer.
		requireCidConfig(t, fapi, cid, &config)
	})
}

func requireFilUnstored(t *testing.T, ctx context.Context, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c)
	require.NoError(t, err)
	require.Empty(t, offers)
}

func requireFilStored(t *testing.T, ctx context.Context, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c)
	require.NoError(t, err)
	require.NotEmpty(t, offers)
}

func TestUnfreeze(t *testing.T) {
	ipfsAPI, fapi, cls := newAPI(t, 1)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	ctx := context.Background()
	cid, data := addRandomFile(t, ra, ipfsAPI)

	config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false).WithHotAllowUnfreeze(true)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	_, err = fapi.Get(ctx, cid)
	require.Equal(t, ffs.ErrHotStorageDisabled, err)

	err = ipfsAPI.Dag().Remove(ctx, cid)
	require.Nil(t, err)
	config = config.WithHotEnabled(true)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	r, err := fapi.Get(ctx, cid)
	require.Nil(t, err)
	fetched, err := ioutil.ReadAll(r)
	require.Nil(t, err)
	require.True(t, bytes.Equal(data, fetched))
}

func TestRenew(t *testing.T) {
	// ToDo: unskip when testnet/3  allows more than one deal
	// See https://bit.ly/2JxQSQk
	t.SkipNow()
	util.AvgBlockTime = time.Millisecond * 200
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })
	ds := tests.NewTxMapDatastore()
	addr, client, ms := newDevnet(t, 2)
	ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	defer closeInternal()

	ra := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, ra, ipfsAPI)

	renewThreshold := 50
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(int64(200)).WithColdFilRenew(true, renewThreshold)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	i, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 1, len(i.Cold.Filecoin.Proposals))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lchain := lotuschain.New(client)
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.Nil(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		h, err := lchain.GetHeight(context.Background())
		require.Nil(t, err)
		if firstDeal.ActivationEpoch+firstDeal.Duration-int64(renewThreshold)+int64(100) > int64(h) {
			require.LessOrEqual(t, len(i.Cold.Filecoin.Proposals), 2)
			continue
		}

		require.Equal(t, len(i.Cold.Filecoin.Proposals), 2)
		require.True(t, firstDeal.Renewed)

		newDeal := i.Cold.Filecoin.Proposals[1]
		require.NotEqual(t, firstDeal.ProposalCid, newDeal.ProposalCid)
		require.False(t, newDeal.Renewed)
		require.Greater(t, newDeal.ActivationEpoch, firstDeal.ActivationEpoch)
		require.Equal(t, config.Cold.Filecoin.DealDuration, newDeal.Duration)
		break Loop
	}
}

func TestRenewWithDecreasedRepFactor(t *testing.T) {
	// ToDo: unskip when testnet/3  allows more than one deal
	// See https://bit.ly/2JxQSQk
	t.SkipNow()
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })
	ds := tests.NewTxMapDatastore()
	addr, client, ms := newDevnet(t, 2)
	ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	defer closeInternal()

	ra := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, ra, ipfsAPI)

	renewThreshold := 50
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(int64(200)).WithColdFilRenew(true, renewThreshold).WithColdFilRepFactor(2)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	// Now decrease RepFactor to 1, so the renewal should consider this.
	// Both now active deals shouldn't be renewed, only one of them.
	config = config.WithColdFilRepFactor(1)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lchain := lotuschain.New(client)
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.Nil(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		secondDeal := i.Cold.Filecoin.Proposals[1]
		h, err := lchain.GetHeight(context.Background())
		require.Nil(t, err)
		if firstDeal.ActivationEpoch+firstDeal.Duration-int64(renewThreshold)+int64(100) > int64(h) {
			require.LessOrEqual(t, len(i.Cold.Filecoin.Proposals), 3)
			continue
		}

		require.Equal(t, len(i.Cold.Filecoin.Proposals), 3)
		// Only one of the two deas should be renewed
		require.True(t, (firstDeal.Renewed && !secondDeal.Renewed) || (secondDeal.Renewed && !firstDeal.Renewed))

		newDeal := i.Cold.Filecoin.Proposals[3]
		require.NotEqual(t, firstDeal.ProposalCid, newDeal.ProposalCid)
		require.False(t, newDeal.Renewed)
		require.Greater(t, newDeal.ActivationEpoch, firstDeal.ActivationEpoch)
		require.Equal(t, config.Cold.Filecoin.DealDuration, newDeal.Duration)
		break Loop
	}
}

func newAPI(t *testing.T, numMiners int) (*httpapi.HttpApi, *api.API, func()) {
	ipfsDocker, cls := tests.LaunchDocker()
	t.Cleanup(func() { cls() })
	ds := tests.NewTxMapDatastore()
	addr, client, ms := newDevnet(t, numMiners)
	ipfsAPI, fapi, closeInternal := newAPIFromDs(t, ds, ffs.EmptyInstanceID, client, addr, ms, ipfsDocker)
	return ipfsAPI, fapi, func() {
		closeInternal()
	}
}

func newDevnet(t *testing.T, numMiners int) (address.Address, *apistruct.FullNodeStruct, ffs.MinerSelector) {
	client, addr, _ := tests.CreateLocalDevnet(t, numMiners)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 1000+i)
	}

	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: "China", EpochPrice: 1000000}
	}
	ms := fixed.New(fixedMiners)
	return addr, client, ms
}

func newAPIFromDs(t *testing.T, ds datastore.TxnDatastore, iid ffs.ApiID, client *apistruct.FullNodeStruct, waddr address.Address, ms ffs.MinerSelector, ipfsDocker *dockertest.Resource) (*httpapi.HttpApi, *api.API, func()) {
	ctx := context.Background()
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfsClient, err := httpapi.NewApi(ipfsAddr)
	require.Nil(t, err)

	dm, err := deals.New(client, deals.WithImportPath(tmpDir))
	require.Nil(t, err)

	fchain := lotuschain.New(client)
	l := cidlogger.New(txndstr.Wrap(ds, "ffs/scheduler/logger"))
	cl := filcold.New(ms, dm, ipfsClient.Dag(), fchain, l)
	cis := cistore.New(txndstr.Wrap(ds, "ffs/scheduler/cistore"))
	pcs := pcstore.New(txndstr.Wrap(ds, "ffs/scheduler/pcstore"))
	js := jstore.New(txndstr.Wrap(ds, "ffs/scheduler/jstore"))
	hl := coreipfs.New(ipfsClient, l)
	sched := scheduler.New(js, pcs, cis, l, hl, cl)

	wm, err := wallet.New(client, &waddr, *big.NewInt(4000000000))
	require.Nil(t, err)

	var fapi *api.API
	if iid == ffs.EmptyInstanceID {
		iid = ffs.NewApiID()
		is := istore.New(iid, txndstr.Wrap(ds, "ffs/api/istore"))
		defConfig := ffs.DefaultCidConfig{
			Hot: ffs.HotConfig{
				Enabled:       true,
				AllowUnfreeze: false,
				Ipfs: ffs.IpfsConfig{
					AddTimeout: 10,
				},
			},
			Cold: ffs.ColdConfig{
				Enabled: true,
				Filecoin: ffs.FilConfig{
					ExcludedMiners: nil,
					DealDuration:   1000,
					RepFactor:      1,
				},
			},
		}
		fapi, err = api.New(ctx, iid, is, sched, wm, defConfig)
		require.Nil(t, err)
	} else {
		is := istore.New(iid, txndstr.Wrap(ds, "ffs/api/istore"))
		fapi, err = api.Load(iid, is, sched, wm)
		require.Nil(t, err)
	}
	time.Sleep(time.Second * 2)

	return ipfsClient, fapi, func() {
		if err := fapi.Close(); err != nil {
			t.Fatalf("closing api: %s", err)
		}
		if err := sched.Close(); err != nil {
			t.Fatalf("closing scheduler: %s", err)
		}
		if err := js.Close(); err != nil {
			t.Fatalf("closing jobstore: %s", err)
		}
	}
}

func requireJobState(t *testing.T, fapi *api.API, jid ffs.JobID, status ffs.JobStatus) ffs.Job {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	ch, err := fapi.Watch(jid)
	require.Nil(t, err)
	defer fapi.Unwatch(ch)
	stop := false
	var res ffs.Job
	for !stop {
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for job update timeout")
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == ffs.Queued || job.Status == ffs.InProgress {
				continue
			}
			require.Equal(t, status, job.Status, job.ErrCause)
			stop = true
			res = job
		}
	}
	fapi.Unwatch(ch)
	return res
}

func requireCidConfig(t *testing.T, fapi *api.API, c cid.Cid, config *ffs.CidConfig) {
	if config == nil {
		defConfig := fapi.GetDefaultCidConfig(c)
		config = &defConfig
	}
	currentConfig, err := fapi.GetCidConfig(c)
	require.NoError(t, err)
	require.Equal(t, *config, currentConfig)
}

func randomBytes(r *rand.Rand, size int) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}

func addRandomFile(t *testing.T, r *rand.Rand, ipfs *httpapi.HttpApi) (cid.Cid, []byte) {
	t.Helper()
	data := randomBytes(r, 600)
	node, err := ipfs.Unixfs().Add(context.Background(), ipfsfiles.NewReaderFile(bytes.NewReader(data)), options.Unixfs.Pin(false))
	if err != nil {
		t.Fatalf("error adding random file: %s", err)
	}
	return node.Cid(), data
}
