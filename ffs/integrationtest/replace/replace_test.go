package replace

import (
	"context"
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

func TestPushCidReplace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ipfs, client, fapi, cls := it.NewAPI(t, 1)
	defer cls()

	r := rand.New(rand.NewSource(22))
	c1, _ := it.AddRandomFile(t, r, ipfs)

	// Test case that an unknown cid is being replaced
	nc, _ := util.CidFromString("Qmc5gCcjYypU7y28oCALwfSvxCBskLuPKWpK4qpterKC7z")
	_, err := fapi.Replace(nc, c1)
	require.Equal(t, api.ErrReplacedCidNotFound, err)

	// Test tipical case
	config := fapi.DefaultStorageConfig().WithColdEnabled(false)
	jid, err := fapi.PushStorageConfig(c1, api.WithStorageConfig(config))
	require.NoError(t, err)
	it.RequireJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, c1, &config)

	c2, _ := it.AddRandomFile(t, r, ipfs)
	jid, err = fapi.Replace(c1, c2)
	require.NoError(t, err)
	it.RequireJobState(t, fapi, jid, ffs.Success)

	config2, err := fapi.GetStorageConfig(c2)
	require.NoError(t, err)
	require.Equal(t, config.Cold.Enabled, config2.Cold.Enabled)

	_, err = fapi.GetStorageConfig(c1)
	require.Equal(t, api.ErrNotFound, err)

	it.RequireIpfsUnpinnedCid(ctx, t, c1, ipfs)
	it.RequireIpfsPinnedCid(ctx, t, c2, ipfs)
	it.RequireFilUnstored(ctx, t, client, c1)
	it.RequireFilUnstored(ctx, t, client, c2)
}

func TestDoubleReplace(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ipfs, ipfsMAddr := it.CreateIPFS(t)
	addr, client, ms := it.NewDevnet(t, 1, ipfsMAddr)
	m, closeManager := it.NewFFSManager(t, ds, client, addr, ms, ipfs)
	defer closeManager()

	// This will ask for a new API to the manager, and do a replace. Always the same replace.
	testAddThenReplace := func() {
		_, authToken, err := m.Create(context.Background())
		require.NoError(t, err)
		time.Sleep(time.Second * 2) // Give some time to the funding transaction

		fapi, err := m.GetByAuthToken(authToken)
		require.NoError(t, err)

		r := rand.New(rand.NewSource(22))
		c1, _ := it.AddRandomFile(t, r, ipfs)
		config := fapi.DefaultStorageConfig().WithColdEnabled(false)
		jid, err := fapi.PushStorageConfig(c1, api.WithStorageConfig(config))
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
		it.RequireStorageConfig(t, fapi, c1, &config)

		c2, _ := it.AddRandomFile(t, r, ipfs)
		jid, err = fapi.Replace(c1, c2)
		require.NoError(t, err)
		it.RequireJobState(t, fapi, jid, ffs.Success)
	}

	// Test the same workflow in different APIs instaneces,
	// but same hot & cold layer.
	testAddThenReplace()
	testAddThenReplace()
}
