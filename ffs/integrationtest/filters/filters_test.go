package filters

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
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/ffs/minerselector/fixed"
	"github.com/textileio/powergate/tests"

	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestFilecoinExcludedMiners(t *testing.T) {
	t.Parallel()
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 2)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		excludedMiner := "f01000"
		config := fapi.DefaultStorageConfig().WithColdFilExcludedMiners([]string{excludedMiner})

		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		p := cinfo.Cold.Filecoin.Proposals[0]
		require.NotEqual(t, p.Miner, excludedMiner)
	})
}

func TestFilecoinTrustedMiner(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 2)
		defer cls()

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		trustedMiner := "f01001"
		config := fapi.DefaultStorageConfig().WithColdFilTrustedMiners([]string{trustedMiner})

		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		p := cinfo.Cold.Filecoin.Proposals[0]
		require.Equal(t, p.Miner, trustedMiner)
	})
}

func TestFilecoinCountryFilter(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfs, ipfsAddr := it.CreateIPFS(t)
		countries := []string{"China", "Uruguay"}
		numMiners := len(countries)
		client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, numMiners, ipfsAddr, false)
		addrs := make([]string, numMiners)
		for i := 0; i < numMiners; i++ {
			addrs[i] = fmt.Sprintf("f0%d", 1000+i)
		}
		fixedMiners := make([]fixed.Miner, len(addrs))
		for i, a := range addrs {
			fixedMiners[i] = fixed.Miner{Addr: a, Country: countries[i], EpochPrice: 500000000}
		}
		ms := fixed.New(fixedMiners)
		ds := tests.NewTxMapDatastore()
		manager, closeManager := itmanager.NewFFSManager(t, ds, client, addr, ms, ipfs)
		defer closeManager()
		auth, err := manager.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3)
		fapi, err := manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)
		countryFilter := []string{"Uruguay"}
		config := fapi.DefaultStorageConfig().WithColdFilCountryCodes(countryFilter)

		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		p := cinfo.Cold.Filecoin.Proposals[0]
		require.Equal(t, p.Miner, "f01001")
	})
}

func TestFilecoinMaxPriceFilter(t *testing.T) {
	t.Parallel()

	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfs, ipfsMAddr := it.CreateIPFS(t)
		client, addr, _ := tests.CreateLocalDevnetWithIPFS(t, 1, ipfsMAddr, false)
		miner := fixed.Miner{Addr: "f01000", EpochPrice: 500000000}
		ms := fixed.New([]fixed.Miner{miner})
		ds := tests.NewTxMapDatastore()
		manager, closeManager := itmanager.NewFFSManager(t, ds, client, addr, ms, ipfs)
		defer closeManager()
		auth, err := manager.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 3) // Wait for funding txn.
		fapi, err := manager.GetByAuthToken(auth.Token)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))
		cid, _ := it.AddRandomFile(t, r, ipfs)

		config := fapi.DefaultStorageConfig().WithColdMaxPrice(400000000)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Failed)

		config = fapi.DefaultStorageConfig().WithColdMaxPrice(600000000)
		jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		p := cinfo.Cold.Filecoin.Proposals[0]
		require.Equal(t, p.Miner, "f01000")
	})
}
