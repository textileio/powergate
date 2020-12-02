package repfactor

import (
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
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestRepFactor(t *testing.T) {
	t.Parallel()
	rfs := []int{1, 2}
	for _, rf := range rfs {
		rf := rf
		t.Run(fmt.Sprintf("%d", rf), func(t *testing.T) {
			t.Parallel()
			tests.RunFlaky(t, func(t *tests.FlakyT) {
				r := rand.New(rand.NewSource(22))
				ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, rf)
				defer cls()
				cid, _ := it.AddRandomFile(t, r, ipfsAPI)
				config := fapi.DefaultStorageConfig().WithColdFilRepFactor(rf)
				jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
				require.NoError(t, err)
				it.RequireEventualJobState(t, fapi, jid, ffs.Success)
				it.RequireStorageConfig(t, fapi, cid, &config)

				cinfo, err := fapi.Show(cid)
				require.NoError(t, err)
				require.Equal(t, rf, len(cinfo.Cold.Filecoin.Proposals))
			})
		})
	}
}

func TestRepFactorIncrease(t *testing.T) {
	t.SkipNow()
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		r := rand.New(rand.NewSource(22))
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 2)
		defer cls()
		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		jid, err := fapi.PushStorageConfig(cid)
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, nil)

		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		require.Equal(t, 1, len(cinfo.Cold.Filecoin.Proposals))
		firstProposal := cinfo.Cold.Filecoin.Proposals[0]

		config := fapi.DefaultStorageConfig().WithColdFilRepFactor(2)
		jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)
		cinfo, err = fapi.Show(cid)
		require.NoError(t, err)
		require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))
		require.Contains(t, cinfo.Cold.Filecoin.Proposals, firstProposal)
	})
}

func TestRepFactorDecrease(t *testing.T) {
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		r := rand.New(rand.NewSource(22))
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 2)
		defer cls()

		cid, _ := it.AddRandomFile(t, r, ipfsAPI)
		config := fapi.DefaultStorageConfig().WithColdFilRepFactor(2)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)

		cinfo, err := fapi.Show(cid)
		require.NoError(t, err)
		require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))

		config = fapi.DefaultStorageConfig().WithColdFilRepFactor(1)
		jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)

		cinfo, err = fapi.Show(cid)
		require.NoError(t, err)
		require.Equal(t, 2, len(cinfo.Cold.Filecoin.Proposals))
	})
}

func TestRenewWithDecreasedRepFactor(t *testing.T) {
	// Too flaky for CI.
	t.SkipNow()
	ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 2)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	cid, _ := it.AddRandomFile(t, ra, ipfsAPI)

	renewThreshold := 12600
	config := fapi.DefaultStorageConfig().WithColdFilDealDuration(int64(100)).WithColdFilRenew(true, renewThreshold).WithColdFilRepFactor(2)
	jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, cid, &config)

	// Now decrease RepFactor to 1, so the renewal should consider this.
	// Both now active deals shouldn't be renewed, only one of them.
	config = config.WithColdFilRepFactor(1)
	jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, cid, &config)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	epochDeadline := 200
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.NoError(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		secondDeal := i.Cold.Filecoin.Proposals[1]
		require.NoError(t, err)
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
