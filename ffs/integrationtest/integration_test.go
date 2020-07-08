package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/deals"
	dealsModule "github.com/textileio/powergate/deals/module"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/cidlogger"
	"github.com/textileio/powergate/ffs/coreipfs"
	"github.com/textileio/powergate/ffs/filcold"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/filchain"
	paych "github.com/textileio/powergate/paych/lotus"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
	"github.com/textileio/powergate/util"
	walletModule "github.com/textileio/powergate/wallet/module"
)

const (
	iWalletBal int64 = 4000000000000000
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	//logging.SetLogLevel("ffs-scheduler", "debug")
	//logging.SetLogLevel("ffs-cidlogger", "debug")

	os.Exit(m.Run())
}

func TestSetDefaultConfig(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
	defer cls()

	config := ffs.DefaultConfig{
		Hot: ffs.HotConfig{
			Enabled: false,
			Ipfs: ffs.IpfsConfig{
				AddTimeout: 10,
			},
		},
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				DealMinDuration: 22333,
				RepFactor:       23,
				Addr:            "123456",
			},
		},
	}
	err := fapi.SetDefaultConfig(config)
	require.Nil(t, err)
	newConfig := fapi.GetDefaultCidConfig(cid.Undef)
	require.Equal(t, newConfig.Hot, config.Hot)
	require.Equal(t, newConfig.Cold, config.Cold)
}

func TestAddrs(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
	defer cls()

	addrs := fapi.Addrs()
	require.Len(t, addrs, 1)
	require.NotEmpty(t, addrs[0].Name)
	require.NotEmpty(t, addrs[0].Addr)
}

func TestNewAddress(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
	defer cls()

	addr, err := fapi.NewAddr(context.Background(), "my address")
	require.Nil(t, err)
	require.NotEmpty(t, addr)

	addrs := fapi.Addrs()
	require.Len(t, addrs, 2)
}

func TestNewAddressDefault(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
	defer cls()

	addr, err := fapi.NewAddr(context.Background(), "my address", api.WithMakeDefault(true))
	require.Nil(t, err)
	require.NotEmpty(t, addr)

	defaultConf := fapi.DefaultConfig()
	require.Equal(t, defaultConf.Cold.Filecoin.Addr, addr)
}

func TestGetDefaultConfig(t *testing.T) {
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
	defer cls()

	defaultConf := fapi.DefaultConfig()
	require.Nil(t, defaultConf.Validate())
}

func TestAdd(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(22))
	t.Run("WithDefaultConfig", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, client, fapi, cls := newAPI(t, 1)
		defer cls()

		cid, _ := addRandomFile(t, r, ipfsAPI)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)
		requireFilStored(ctx, t, client, cid)
		requireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
		requireStorageDealRecord(t, fapi, cid)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		ipfsAPI, _, fapi, cls := newAPI(t, 1)
		defer cls()
		cid, _ := addRandomFile(t, r, ipfsAPI)

		config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false).WithColdFilDealDuration(int64(1234))
		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireStorageDealRecord(t, fapi, cid)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, _, fapi, cls := newAPI(t, 1)
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
	t.Parallel()
	ctx := context.Background()
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	var err error
	var first api.InstanceInfo
	t.Run("Minimal", func(t *testing.T) {
		first, err = fapi.Info(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, first.ID)
		require.Len(t, first.Balances, 1)
		require.NotEmpty(t, first.Balances[0].Addr)
		require.Greater(t, first.Balances[0].Balance, uint64(0))
		require.Equal(t, len(first.Pins), 0)
	})

	r := rand.New(rand.NewSource(22))
	n := 3
	for i := 0; i < n; i++ {
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)
	}

	t.Run("WithThreeAdd", func(t *testing.T) {
		second, err := fapi.Info(ctx)
		require.Nil(t, err)
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
	ipfs, _, fapi, cls := newAPI(t, 1)

	defer cls()

	t.Run("NotStored", func(t *testing.T) {
		c, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
		_, err := fapi.Show(c)
		require.Equal(t, api.ErrNotFound, err)
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
		require.Greater(t, p.EpochPrice, uint64(0))
	})
}

func TestColdInstanceLoad(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := createIPFS(t)
	addr, client, ms := newDevnet(t, 1, ipfsMAddr)
	manager, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)

	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.

	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)

	ra := rand.New(rand.NewSource(22))
	cid, data := addRandomFile(t, ra, ipfs)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, nil)

	info, err := fapi.Info(ctx)
	require.Nil(t, err)
	shw, err := fapi.Show(cid)
	require.Nil(t, err)

	// Now close the FFS Instance, and the manager.
	err = fapi.Close()
	require.NoError(t, err)
	closeManager()

	// Rehydrate things again and check state.
	manager, closeManager = newFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()
	fapi, err = manager.GetByAuthToken(auth)
	require.NoError(t, err)

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
	t.Parallel()
	rfs := []int{1, 2}
	r := rand.New(rand.NewSource(22))
	for _, rf := range rfs {
		t.Run(fmt.Sprintf("%d", rf), func(t *testing.T) {
			ipfsAPI, _, fapi, cls := newAPI(t, rf)
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
	t.Parallel()
	r := rand.New(rand.NewSource(22))
	ipfsAPI, _, fapi, cls := newAPI(t, 2)
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
	t.Parallel()
	r := rand.New(rand.NewSource(22))
	ipfsAPI, _, fapi, cls := newAPI(t, 2)
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
	t.Parallel()
	_, _, fapi, cls := newAPI(t, 1)
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
	t.Parallel()
	ipfsAPI, _, fapi, cls := newAPI(t, 1)
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
	require.Greater(t, p.Duration, duration)
	require.Greater(t, p.ActivationEpoch, int64(0))
}

func TestFilecoinExcludedMiners(t *testing.T) {
	t.Parallel()
	ipfsAPI, _, fapi, cls := newAPI(t, 2)
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

func TestFilecoinTrustedMiner(t *testing.T) {
	t.Parallel()
	ipfsAPI, _, fapi, cls := newAPI(t, 2)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfsAPI)
	trustedMiner := "t01001"
	config := fapi.GetDefaultCidConfig(cid).WithColdFilTrustedMiners([]string{trustedMiner})

	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, p.Miner, trustedMiner)
}

func TestFilecoinCountryFilter(t *testing.T) {
	t.Parallel()
	ipfs, ipfsAddr := createIPFS(t)
	countries := []string{"China", "Uruguay"}
	numMiners := len(countries)
	client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, numMiners, ipfsAddr, false)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 1000+i)
	}
	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: countries[i], EpochPrice: 500000000}
	}
	ms := fixed.New(fixedMiners)
	ds := tests.NewTxMapDatastore()
	manager, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()
	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3)
	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfs)
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

func TestFilecoinMaxPriceFilter(t *testing.T) {
	t.Parallel()
	ipfs, ipfsMAddr := createIPFS(t)
	client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, 1, ipfsMAddr, false)
	miner := fixed.Miner{Addr: "t01000", EpochPrice: 500000000}
	ms := fixed.New([]fixed.Miner{miner})
	ds := tests.NewTxMapDatastore()
	manager, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()
	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn.
	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfs)

	config := fapi.GetDefaultCidConfig(cid).WithColdMaxPrice(400000000)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Failed)

	config = fapi.GetDefaultCidConfig(cid).WithColdMaxPrice(600000000)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)
	cinfo, err := fapi.Show(cid)
	require.Nil(t, err)
	p := cinfo.Cold.Filecoin.Proposals[0]
	require.Equal(t, p.Miner, "t01000")
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
		name := fmt.Sprintf("Hot(%v)/Cold(%v)", tt.HotEnabled, tt.ColdEnabled)
		t.Run(name, func(t *testing.T) {
			ipfsAPI, _, fapi, cls := newAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := addRandomFile(t, r, ipfsAPI)
			config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(tt.ColdEnabled).WithHotEnabled(tt.HotEnabled)

			jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
			require.Nil(t, err)

			expectedJobState := ffs.Success
			requireJobState(t, fapi, jid, expectedJobState)

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
					expectedErr = api.ErrHotStorageDisabled
				}
				require.Equal(t, expectedErr, err)

				// External assertions
				if !tt.HotEnabled {
					requireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)
				} else {
					requireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
				}
			}
		})
	}
}

func requireIpfsUnpinnedCid(ctx context.Context, t *testing.T, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
	pins, err := ipfsAPI.Pin().Ls(ctx)
	require.NoError(t, err)
	for _, p := range pins {
		require.NotEqual(t, cid, p.Path().Cid(), "Cid isn't unpined from IPFS node")
	}
}

func requireIpfsPinnedCid(ctx context.Context, t *testing.T, cid cid.Cid, ipfsAPI *httpapi.HttpApi) {
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
	t.Parallel()
	t.Run("HotEnabledDisabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, _, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsPinnedCid(ctx, t, cid, ipfsAPI)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(false)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)
	})
	t.Run("HotDisabledEnabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, _, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithHotEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsUnpinnedCid(ctx, t, cid, ipfsAPI)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireIpfsPinnedCid(ctx, t, cid, ipfsAPI)
	})
	t.Run("ColdDisabledEnabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, client, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilUnstored(ctx, t, client, cid)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilStored(ctx, t, client, cid)
	})
	t.Run("ColdEnabledDisabled", func(t *testing.T) {
		ctx := context.Background()
		ipfsAPI, client, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfsAPI)
		config := fapi.GetDefaultCidConfig(cid).WithColdEnabled(false)

		jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, &config)
		requireFilUnstored(ctx, t, client, cid)

		config = fapi.GetDefaultCidConfig(cid).WithHotEnabled(true)
		jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)

		// Yes, still stored in filecoin since deals can't be
		// undone.
		requireFilStored(ctx, t, client, cid)
		// Despite of the above, check that the Cid Config still reflects
		// that this *shouldn't* be in the Cold Storage. To indicate
		// this can't be renewed, or any other future action that tries to
		// store it again in the Cold Layer.
		requireCidConfig(t, fapi, cid, &config)
	})
}

func requireFilUnstored(ctx context.Context, t *testing.T, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c)
	require.NoError(t, err)
	require.Empty(t, offers)
}

func requireFilStored(ctx context.Context, t *testing.T, client *apistruct.FullNodeStruct, c cid.Cid) {
	offers, err := client.ClientFindData(ctx, c)
	require.NoError(t, err)
	require.NotEmpty(t, offers)
}

func TestUnfreeze(t *testing.T) {
	t.Parallel()
	ipfsAPI, _, fapi, cls := newAPI(t, 1)
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
	require.Equal(t, api.ErrHotStorageDisabled, err)

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
	requireRetrievalDealRecord(t, fapi, cid)
}

func TestRenew(t *testing.T) {
	t.Parallel()
	ipfsAPI, _, fapi, cls := newAPI(t, 1)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, ra, ipfsAPI)

	renewThreshold := 12600
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(int64(100)).WithColdFilRenew(true, renewThreshold)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	i, err := fapi.Show(cid)
	require.Nil(t, err)
	require.Equal(t, 1, len(i.Cold.Filecoin.Proposals))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	epochDeadline := 200
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.Nil(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		require.Nil(t, err)
		if len(i.Cold.Filecoin.Proposals) < 2 {
			require.Greater(t, epochDeadline, 0)
			continue
		}

		require.Equal(t, 2, len(i.Cold.Filecoin.Proposals))
		require.True(t, firstDeal.Renewed)

		newDeal := i.Cold.Filecoin.Proposals[1]
		require.NotEqual(t, firstDeal.ProposalCid, newDeal.ProposalCid)
		require.False(t, newDeal.Renewed)
		require.Greater(t, newDeal.ActivationEpoch, firstDeal.ActivationEpoch)
		require.Greater(t, newDeal.Duration, config.Cold.Filecoin.DealMinDuration)
		break Loop
	}
}

func TestRenewWithDecreasedRepFactor(t *testing.T) {
	// Too flaky for CI.
	t.SkipNow()
	ipfsAPI, _, fapi, cls := newAPI(t, 2)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, ra, ipfsAPI)

	renewThreshold := 12600
	config := fapi.GetDefaultCidConfig(cid).WithColdFilDealDuration(int64(100)).WithColdFilRenew(true, renewThreshold).WithColdFilRepFactor(2)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	// Now decrease RepFactor to 1, so the renewal should consider this.
	// Both now active deals shouldn't be renewed, only one of them.
	config = config.WithColdFilRepFactor(1)
	jid, err = fapi.PushConfig(cid, api.WithCidConfig(config), api.WithOverride(true))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	epochDeadline := 200
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.Nil(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		secondDeal := i.Cold.Filecoin.Proposals[1]
		require.Nil(t, err)
		if len(i.Cold.Filecoin.Proposals) < 3 {
			epochDeadline--
			require.Greater(t, epochDeadline, 0)
			continue
		}

		require.Equal(t, 3, len(i.Cold.Filecoin.Proposals))
		// Only one of the two deas should be renewed
		require.True(t, (firstDeal.Renewed && !secondDeal.Renewed) || (secondDeal.Renewed && !firstDeal.Renewed))

		newDeal := i.Cold.Filecoin.Proposals[2]
		require.NotEqual(t, firstDeal.ProposalCid, newDeal.ProposalCid)
		require.False(t, newDeal.Renewed)
		require.Greater(t, newDeal.ActivationEpoch, firstDeal.ActivationEpoch)
		require.Greater(t, newDeal.Duration, config.Cold.Filecoin.DealMinDuration)
		break Loop
	}
}

func TestCidLogger(t *testing.T) {
	t.Parallel()
	t.Run("WithNoFilters", func(t *testing.T) {
		ipfs, _, fapi, cls := newAPI(t, 1)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch := make(chan ffs.LogEntry)
		go func() {
			err = fapi.WatchLogs(ctx, ch, cid)
			close(ch)
		}()
		stop := false
		for !stop {
			select {
			case le, ok := <-ch:
				if !ok {
					require.NoError(t, err)
					stop = true
					continue
				}
				cancel()
				require.Equal(t, cid, le.Cid)
				require.Equal(t, jid, le.Jid)
				require.True(t, time.Since(le.Timestamp) < time.Second*5)
				require.NotEmpty(t, le.Msg)
			case <-time.After(time.Second):
				t.Fatal("no cid logs were received")
			}
		}

		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, cid, nil)
	})
	t.Run("WithJidFilter", func(t *testing.T) {
		t.Run("CorrectJid", func(t *testing.T) {
			ipfs, _, fapi, cls := newAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := addRandomFile(t, r, ipfs)
			jid, err := fapi.PushConfig(cid)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := make(chan ffs.LogEntry)
			go func() {
				err = fapi.WatchLogs(ctx, ch, cid, api.WithJidFilter(jid))
				close(ch)
			}()
			stop := false
			for !stop {
				select {
				case le, ok := <-ch:
					if !ok {
						require.NoError(t, err)
						stop = true
						continue
					}
					cancel()
					require.Equal(t, cid, le.Cid)
					require.Equal(t, jid, le.Jid)
					require.True(t, time.Since(le.Timestamp) < time.Second*5)
					require.NotEmpty(t, le.Msg)
				case <-time.After(time.Second):
					t.Fatal("no cid logs were received")
				}
			}

			requireJobState(t, fapi, jid, ffs.Success)
			requireCidConfig(t, fapi, cid, nil)
		})
		t.Run("IncorrectJid", func(t *testing.T) {
			ipfs, _, fapi, cls := newAPI(t, 1)
			defer cls()

			r := rand.New(rand.NewSource(22))
			cid, _ := addRandomFile(t, r, ipfs)
			jid, err := fapi.PushConfig(cid)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := make(chan ffs.LogEntry)
			go func() {
				fakeJid := ffs.NewJobID()
				err = fapi.WatchLogs(ctx, ch, cid, api.WithJidFilter(fakeJid))
				close(ch)
			}()
			select {
			case <-ch:
				t.Fatal("the channels shouldn't receive any log messages")
			case <-time.After(3 * time.Second):
			}
			require.NoError(t, err)

			requireJobState(t, fapi, jid, ffs.Success)
			requireCidConfig(t, fapi, cid, nil)
		})
	})
}

func TestPushCidReplace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, client, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c1, _ := addRandomFile(t, r, ipfs)

	// Test case that an unknown cid is being replaced
	nc, _ := cid.Decode("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
	_, err := fapi.Replace(nc, c1)
	require.Equal(t, api.ErrReplacedCidNotFound, err)

	// Test tipical case
	config := fapi.GetDefaultCidConfig(c1).WithColdEnabled(false)
	jid, err := fapi.PushConfig(c1, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, c1, &config)

	c2, _ := addRandomFile(t, r, ipfs)
	jid, err = fapi.Replace(c1, c2)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)

	config2, err := fapi.GetCidConfig(c2)
	require.NoError(t, err)
	require.Equal(t, config.Cold.Enabled, config2.Cold.Enabled)

	_, err = fapi.GetCidConfig(c1)
	require.Equal(t, api.ErrNotFound, err)

	requireIpfsUnpinnedCid(ctx, t, c1, ipfs)
	requireIpfsPinnedCid(ctx, t, c2, ipfs)
	requireFilUnstored(ctx, t, client, c1)
	requireFilUnstored(ctx, t, client, c2)
}

func TestDoubleReplace(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := createIPFS(t)
	addr, client, ms := newDevnet(t, 1, ipfsMAddr)
	m, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()

	// This will ask for a new API to the manager, and do a replace. Always the same replace.
	testAddThenReplace := func() {
		_, authToken, err := m.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 2) // Give some time to the funding transaction

		fapi, err := m.GetByAuthToken(authToken)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))
		c1, _ := addRandomFile(t, r, ipfs)
		config := fapi.GetDefaultCidConfig(c1).WithColdEnabled(false)
		jid, err := fapi.PushConfig(c1, api.WithCidConfig(config))
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
		requireCidConfig(t, fapi, c1, &config)

		c2, _ := addRandomFile(t, r, ipfs)
		jid, err = fapi.Replace(c1, c2)
		require.Nil(t, err)
		requireJobState(t, fapi, jid, ffs.Success)
	}

	// Test the same workflow in different APIs instaneces,
	// but same hot & cold layer.
	testAddThenReplace()
	testAddThenReplace()
}

func TestRemove(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c1, _ := addRandomFile(t, r, ipfs)

	config := fapi.GetDefaultCidConfig(c1).WithColdEnabled(false)
	jid, err := fapi.PushConfig(c1, api.WithCidConfig(config))
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, c1, &config)

	err = fapi.Remove(c1)
	require.Equal(t, api.ErrActiveInStorage, err)

	config = config.WithHotEnabled(false)
	jid, err = fapi.PushConfig(c1, api.WithCidConfig(config), api.WithOverride(true))
	requireJobState(t, fapi, jid, ffs.Success)
	require.Nil(t, err)

	err = fapi.Remove(c1)
	require.Nil(t, err)
	_, err = fapi.GetCidConfig(c1)
	require.Equal(t, api.ErrNotFound, err)
}

func TestSendFil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, _, fapi, cls := newAPI(t, 1)
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
	require.Nil(t, err)

	hasInitialBal := func() bool {
		bal, err := balForAddress(addr2)
		require.Nil(t, err)
		return bal == uint64(iWalletBal)
	}

	hasNewBal := func() bool {
		bal, err := balForAddress(addr2)
		require.Nil(t, err)
		return bal == uint64(iWalletBal+amt)
	}

	require.Eventually(t, hasInitialBal, time.Second*5, time.Second)

	err = fapi.SendFil(ctx, addr1, addr2, big.NewInt(amt))
	require.Nil(t, err)

	require.Eventually(t, hasNewBal, time.Second*5, time.Second)
}

// This isn't very nice way to test for repair. The main problem is that now
// deal start is buffered for future start for 10000 blocks at the Lotus level.
// Se we can't wait that much on a devnet. That setup has some ToDo comments so
// most prob will change and we can do some nicier test here.
// Better than no test is some test, so this tests that the repair logic gets triggered
// and the related Job ran successfully.
func TestRepair(t *testing.T) {
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	cid, _ := addRandomFile(t, r, ipfs)
	config := fapi.GetDefaultCidConfig(cid).WithRepairable(true)
	jid, err := fapi.PushConfig(cid, api.WithCidConfig(config))
	require.NoError(t, err)
	requireJobState(t, fapi, jid, ffs.Success)
	requireCidConfig(t, fapi, cid, &config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan ffs.LogEntry)
	go func() {
		err = fapi.WatchLogs(ctx, ch, cid)
		close(ch)
	}()
	stop := false
	for !stop {
		select {
		case le, ok := <-ch:
			if !ok {
				require.NoError(t, err)
				stop = true
				continue
			}
			// Expected message: "Job %s was queued for repair evaluation."
			if strings.Contains(le.Msg, "was queued for repair evaluation.") {
				parts := strings.SplitN(le.Msg, " ", 3)
				require.Equal(t, 3, len(parts), "Log message is malformed")
				jid := ffs.JobID(parts[1])
				var err2 error
				ctx2, cancel2 := context.WithCancel(context.Background())
				ch := make(chan ffs.Job, 1)
				go func() {
					err2 = fapi.WatchJobs(ctx2, ch, jid)
					close(ch)
				}()
				repairJob := <-ch
				cancel2()
				<-ch
				require.Nil(t, err2)
				requireJobState(t, fapi, repairJob.ID, ffs.Success)
				requireCidConfig(t, fapi, cid, &config)
				cancel()
			}
		case <-time.After(time.Second * 10):
			t.Fatal("no cid logs related with repairing were received")
		}
	}
}

func TestFailedJobMessage(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	// Add a file size that would be bigger than the
	// sector size. This should make the deal fail on the miner.
	c1, _ := addRandomFileSize(t, r, ipfs, 2000)

	jid, err := fapi.PushConfig(c1)
	require.Nil(t, err)
	job := requireJobState(t, fapi, jid, ffs.Failed)
	require.NotEmpty(t, job.ErrCause)
	require.Len(t, job.DealErrors, 1)
	de := job.DealErrors[0]
	require.False(t, de.ProposalCid.Defined())
	require.NotEmpty(t, de.Miner)
	require.Equal(t, "failed to start deal: cannot propose a deal whose piece size (4096) is greater than sector size (2048)", de.Message)
}

func TestLogHistory(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c, _ := addRandomFile(t, r, ipfs)
	jid, err := fapi.PushConfig(c)
	require.Nil(t, err)
	job := requireJobState(t, fapi, jid, ffs.Success)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	ch := make(chan ffs.LogEntry, 100)
	go func() {
		err = fapi.WatchLogs(ctx, ch, c, api.WithHistory(true))
		close(ch)
	}()
	var lgs []ffs.LogEntry
	for le := range ch {
		require.Equal(t, c, le.Cid)
		require.Equal(t, job.ID, le.Jid)
		require.NotEmpty(t, le.Msg)
		if len(lgs) > 0 {
			require.True(t, le.Timestamp.After(lgs[len(lgs)-1].Timestamp))
		}

		lgs = append(lgs, le)
	}
	require.NoError(t, err)
	require.Greater(t, len(lgs), 3) // Ask to have more than 3 log messages.
}

func TestResumeScheduler(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := createIPFS(t)
	addr, client, ms := newDevnet(t, 1, ipfsMAddr)
	manager, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)
	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.
	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)

	r := rand.New(rand.NewSource(22))
	c, _ := addRandomFile(t, r, ipfs)
	jid, err := fapi.PushConfig(c)
	require.Nil(t, err)

	time.Sleep(time.Second * 3)
	ds2, err := ds.Clone()
	require.NoError(t, err)
	closeManager()

	manager, closeManager = newFFSManager(t, ds2, client, addr, ms, ipfs)
	defer closeManager()
	fapi, err = manager.GetByAuthToken(auth) // Get same FFS instance again
	require.NoError(t, err)
	requireJobState(t, fapi, jid, ffs.Success)

	sh, err := fapi.Show(c)
	require.NoError(t, err)
	require.Equal(t, 1, len(sh.Cold.Filecoin.Proposals)) // Check only one deal still exits.
}

func TestParallelExecution(t *testing.T) {
	t.Parallel()
	ipfs, _, fapi, cls := newAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	n := 3
	cids := make([]cid.Cid, n)
	jids := make([]ffs.JobID, n)
	for i := 0; i < n; i++ {
		cid, _ := addRandomFile(t, r, ipfs)
		jid, err := fapi.PushConfig(cid)
		require.Nil(t, err)
		cids[i] = cid
		jids[i] = jid
		// Add some sleep time to avoid all of them
		// being batched in the same scheduler run.
		time.Sleep(time.Millisecond * 100)
	}
	// Check that all jobs should be immediately in the Executing status, since
	// the default max parallel runs is 50. So all should get in.
	for i := 0; i < len(jids); i++ {
		requireJobState(t, fapi, jids[i], ffs.Executing)
	}

	// Now just check that all of them finish successfully.
	for i := 0; i < len(jids); i++ {
		requireJobState(t, fapi, jids[i], ffs.Executing)
		requireCidConfig(t, fapi, cids[i], nil)
	}
}

func TestJobCancellation(t *testing.T) {
	r := rand.New(rand.NewSource(22))
	ipfsAPI, _, fapi, cls := newAPI(t, 1)
	defer cls()

	cid, _ := addRandomFile(t, r, ipfsAPI)
	jid, err := fapi.PushConfig(cid)
	require.Nil(t, err)
	requireJobState(t, fapi, jid, ffs.Executing)
	time.Sleep(time.Second * 2)

	err = fapi.CancelJob(jid)
	require.NoError(t, err)

	// Assert that the Job status is Canceled, *and* was
	// finished _fast_.
	before := time.Now()
	requireJobState(t, fapi, jid, ffs.Canceled)
	require.True(t, time.Since(before) < time.Second)
}

func newAPI(t *testing.T, numMiners int) (*httpapi.HttpApi, *apistruct.FullNodeStruct, *api.API, func()) {
	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := createIPFS(t)
	addr, client, ms := newDevnet(t, numMiners, ipfsMAddr)
	manager, closeManager := newFFSManager(t, ds, client, addr, ms, ipfs)
	_, auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.
	fapi, err := manager.GetByAuthToken(auth)
	require.NoError(t, err)
	return ipfs, client, fapi, func() {
		err := fapi.Close()
		require.NoError(t, err)
		closeManager()
	}
}

func createIPFS(t *testing.T) (*httpapi.HttpApi, string) {
	ipfsDocker, cls := tests.LaunchIPFSDocker()
	t.Cleanup(func() { cls() })
	ipfsAddr := util.MustParseAddr("/ip4/127.0.0.1/tcp/" + ipfsDocker.GetPort("5001/tcp"))
	ipfs, err := httpapi.NewApi(ipfsAddr)
	require.NoError(t, err)
	bridgeIP := ipfsDocker.Container.NetworkSettings.Networks["bridge"].IPAddress
	ipfsDockerMAddr := fmt.Sprintf("/ip4/%s/tcp/5001", bridgeIP)

	return ipfs, ipfsDockerMAddr
}

func newDevnet(t *testing.T, numMiners int, ipfsAddr string) (address.Address, *apistruct.FullNodeStruct, ffs.MinerSelector) {
	client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, numMiners, ipfsAddr, false)
	addrs := make([]string, numMiners)
	for i := 0; i < numMiners; i++ {
		addrs[i] = fmt.Sprintf("t0%d", 1000+i)
	}

	fixedMiners := make([]fixed.Miner, len(addrs))
	for i, a := range addrs {
		fixedMiners[i] = fixed.Miner{Addr: a, Country: "China", EpochPrice: 500000000}
	}
	ms := fixed.New(fixedMiners)
	return addr, client, ms
}

func newFFSManager(t *testing.T, ds datastore.TxnDatastore, lotusClient *apistruct.FullNodeStruct, masterAddr address.Address, ms ffs.MinerSelector, ipfsClient *httpapi.HttpApi) (*manager.Manager, func()) {
	dm, err := dealsModule.New(txndstr.Wrap(ds, "deals"), lotusClient)
	require.Nil(t, err)

	fchain := filchain.New(lotusClient)
	l := cidlogger.New(txndstr.Wrap(ds, "ffs/cidlogger"))
	cl := filcold.New(ms, dm, ipfsClient, fchain, l)
	hl, err := coreipfs.New(ipfsClient, l)
	require.NoError(t, err)
	sched, err := scheduler.New(txndstr.Wrap(ds, "ffs/scheduler"), l, hl, cl)
	require.NoError(t, err)

	wm, err := walletModule.New(lotusClient, masterAddr, *big.NewInt(iWalletBal), false)
	require.Nil(t, err)

	pm := paych.New(lotusClient)

	manager, err := manager.New(ds, wm, pm, dm, sched)
	require.NoError(t, err)
	err = manager.SetDefaultConfig(ffs.DefaultConfig{
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
				ExcludedMiners:  nil,
				DealMinDuration: 1000,
				RepFactor:       1,
			},
		},
	})
	require.NoError(t, err)

	return manager, func() {
		if err := manager.Close(); err != nil {
			t.Fatalf("closing api: %s", err)
		}
		if err := sched.Close(); err != nil {
			t.Fatalf("closing scheduler: %s", err)
		}
		if err := l.Close(); err != nil {
			t.Fatalf("closing cidlogger: %s", err)
		}
	}
}

func requireJobState(t *testing.T, fapi *api.API, jid ffs.JobID, status ffs.JobStatus) ffs.Job {
	t.Helper()
	ch := make(chan ffs.Job)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var err error
	go func() {
		err = fapi.WatchJobs(ctx, ch, jid)
		close(ch)
	}()
	stop := false
	var res ffs.Job
	for !stop {
		select {
		case <-time.After(60 * time.Second):
			t.Fatalf("waiting for job update timeout")
		case job, ok := <-ch:
			require.True(t, ok)
			require.Equal(t, jid, job.ID)
			if job.Status == ffs.Queued || job.Status == ffs.Executing {
				if job.Status == status {
					stop = true
					res = job
				}
				continue
			}
			require.Equal(t, status, job.Status, job.ErrCause)
			stop = true
			res = job
		}
	}
	require.NoError(t, err)
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

func requireStorageDealRecord(t *testing.T, fapi *api.API, c cid.Cid) {
	time.Sleep(time.Second)
	recs, err := fapi.ListStorageDealRecords(deals.WithIncludeFinal(true))
	require.Nil(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].RootCid)
}

func requireRetrievalDealRecord(t *testing.T, fapi *api.API, c cid.Cid) {
	recs, err := fapi.ListRetrievalDealRecords()
	require.Nil(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, c, recs[0].DealInfo.RootCid)
}

func randomBytes(r *rand.Rand, size int) []byte {
	buf := make([]byte, size)
	_, _ = r.Read(buf)
	return buf
}

func addRandomFile(t *testing.T, r *rand.Rand, ipfs *httpapi.HttpApi) (cid.Cid, []byte) {
	return addRandomFileSize(t, r, ipfs, 600)
}

func addRandomFileSize(t *testing.T, r *rand.Rand, ipfs *httpapi.HttpApi, size int) (cid.Cid, []byte) {
	t.Helper()
	data := randomBytes(r, size)
	node, err := ipfs.Unixfs().Add(context.Background(), ipfsfiles.NewReaderFile(bytes.NewReader(data)), options.Unixfs.Pin(false))
	if err != nil {
		t.Fatalf("error adding random file: %s", err)
	}
	return node.Cid(), data
}
