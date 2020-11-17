package unfreeze

import (
	"bytes"
	"context"
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
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestUnfreeze(t *testing.T) {
	t.Parallel()
	scheduler.HardcodedHotTimeout = time.Second * 10
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		ra := rand.New(rand.NewSource(22))
		ctx := context.Background()
		cid, data := it.AddRandomFile(t, ra, ipfsAPI)

		config := fapi.DefaultStorageConfig().WithHotEnabled(false).WithHotAllowUnfreeze(true)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)

		_, err = fapi.Get(ctx, cid)
		require.Equal(t, api.ErrHotStorageDisabled, err)

		err = ipfsAPI.Dag().Remove(ctx, cid)
		require.NoError(t, err)
		config = config.WithHotEnabled(true)
		jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, cid, &config)

		r, err := fapi.Get(ctx, cid)
		require.NoError(t, err)
		fetched, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		require.True(t, bytes.Equal(data, fetched))
		it.RequireRetrievalDealRecord(t, fapi, cid)
	})
}
