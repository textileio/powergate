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
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/api"
	it "github.com/textileio/powergate/v2/ffs/integrationtest"
	itmanager "github.com/textileio/powergate/v2/ffs/integrationtest/manager"
	"github.com/textileio/powergate/v2/tests"
	"github.com/textileio/powergate/v2/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 500
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestUnfreeze(t *testing.T) {
	t.Parallel()
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

func TestUnfreezeRecords(t *testing.T) {
	t.Parallel()
	tests.RunFlaky(t, func(t *tests.FlakyT) {
		ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 1)
		defer cls()

		// Step 1. Produce a Filecoin retrieval.
		ra := rand.New(rand.NewSource(22))
		ctx := context.Background()
		cid, _ := it.AddRandomFile(t, ra, ipfsAPI)

		config := fapi.DefaultStorageConfig().WithHotEnabled(false).WithHotAllowUnfreeze(true)
		jid, err := fapi.PushStorageConfig(cid, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)

		err = ipfsAPI.Dag().Remove(ctx, cid)
		require.NoError(t, err)
		config = config.WithHotEnabled(true)
		jid, err = fapi.PushStorageConfig(cid, api.WithStorageConfig(config), api.WithOverride(true))
		require.NoError(t, err)
		it.RequireEventualJobState(t, fapi, jid, ffs.Success)

		// Step 2. Check that the Retrieval Record has appropriate values.
		recs, err := fapi.RetrievalDealRecords()
		require.NoError(t, err)
		require.Len(t, recs, 1)

		rr := recs[0]
		require.Equal(t, cid, rr.DealInfo.RootCid)
		dtStart := time.Unix(rr.DataTransferStart, 0)
		dtEnd := time.Unix(rr.DataTransferEnd, 0)
		require.False(t, dtStart.IsZero())
		require.False(t, dtEnd.IsZero())
		// Retrievals happen too fast in CI to enforce
		// a strict start < end, so we consider == too.
		// The designed granularity is 'seconds', not 'nanoseconds'.
		require.True(t, dtStart.Before(dtEnd) || dtStart.Equal(dtEnd))
		require.Empty(t, rr.ErrMsg)
	})
}
