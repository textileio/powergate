package renew

import (
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

	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestRenew(t *testing.T) {
	// Unfortunately, a minimum deal duration of 180 *days* is enforced at
	// the Filecoin network. That's a lot of blocks to wait in a CI test.
	// We should eventually have some workaround this in the localnet.
	t.SkipNow()

	t.Parallel()
	ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 1)
	defer cls()

	ra := rand.New(rand.NewSource(22))
	cid, _ := it.AddRandomFile(t, ra, ipfsAPI)

	renewThreshold := 12600
	config := fapi.DefaultStorageConfig().WithColdFilDealDuration(util.MinDealDuration+int64(100)).WithColdFilRenew(true, renewThreshold)
	jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, cid, &config)

	i, err := fapi.Show(cid)
	require.NoError(t, err)
	require.Equal(t, 1, len(i.Cold.Filecoin.Proposals))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	epochDeadline := 200
Loop:
	for range ticker.C {
		i, err := fapi.Show(cid)
		require.NoError(t, err)

		firstDeal := i.Cold.Filecoin.Proposals[0]
		require.NoError(t, err)
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
