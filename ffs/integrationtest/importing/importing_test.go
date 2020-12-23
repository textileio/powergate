package importing

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	it "github.com/textileio/powergate/ffs/integrationtest"
	itmanager "github.com/textileio/powergate/ffs/integrationtest/manager"
	"github.com/textileio/powergate/tests"
)

func TestImportWithoutRetrievalTwoUsers(t *testing.T) {
	t.Parallel()
	ds := tests.NewTxMapDatastore()
	ipfsAPI, ipfsMAddr := it.CreateIPFS(t)
	addr, client, ms := itmanager.NewDevnet(t, 1, ipfsMAddr)
	manager, closeManager := itmanager.NewFFSManager(t, ds, client, addr, ms, ipfsAPI)
	defer closeManager()

	// Store some data with User 1.
	auth, err := manager.Create(context.Background())
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.
	fapi1, err := manager.GetByAuthToken(auth.Token)
	require.NoError(t, err)
	r := rand.New(rand.NewSource(22))
	c, _ := it.AddRandomFile(t, r, ipfsAPI)
	jid, err := fapi1.PushStorageConfig(c)
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi1, jid, ffs.Success)

	// Grab the DealID from the created deal.
	si1, err := fapi1.StorageInfo(c)
	require.NoError(t, err)
	si1p := si1.Cold.Filecoin.Proposals[0]

	// Create some User 2, import the DealIDs from User 1.
	// Make the import without unfreeze, so just import deals
	// and check StorageInfo is what we expect.
	auth, err = manager.Create(context.Background())
	require.NoError(t, err)
	require.NoError(t, err)
	time.Sleep(time.Second * 3) // Wait for funding txn to finish.
	fapi2, err := manager.GetByAuthToken(auth.Token)
	require.NoError(t, err)
	config := fapi2.DefaultStorageConfig().WithHotEnabled(false)
	jid, err = fapi2.PushStorageConfig(c, api.WithDealImport([]uint64{si1p.DealID}), api.WithStorageConfig(config))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi2, jid, ffs.Success)

	// Now compare that User 2 active deals for the Cid matches the same
	// data of User 1. That should be true, since User 2 imported the deal
	// created by User 1.
	si2, err := fapi2.StorageInfo(c)
	require.NoError(t, err)
	require.Equal(t, fapi2.ID(), si2.APIID)
	require.Equal(t, c, si2.Cid)
	require.Equal(t, jid, si2.JobID)
	require.Len(t, si2.Cold.Filecoin.Proposals, 1)
	si2p := si2.Cold.Filecoin.Proposals[0]
	require.Equal(t, si1p.DealID, si2p.DealID)
	require.Equal(t, si1p.PieceCid, si2p.PieceCid)
	require.Equal(t, si1p.Duration, si2p.Duration)
	require.Equal(t, si1p.StartEpoch, si2p.StartEpoch)
	require.Equal(t, si1p.Miner, si2p.Miner)
	require.Equal(t, si1.Cold.Filecoin.Size, si2.Cold.Filecoin.Size)
}

func TestImportWithRetrievalSingleUser(t *testing.T) {
	t.Parallel()

	ipfsAPI, _, fapi, cls := itmanager.NewAPI(t, 1)
	defer cls()

	// Add some data to Filecoin.
	r := rand.New(rand.NewSource(22))
	c, data := it.AddRandomFile(t, r, ipfsAPI)
	cfg := fapi.DefaultStorageConfig().WithHotEnabled(false).WithHotAllowUnfreeze(true)
	jid, err := fapi.PushStorageConfig(c, api.WithStorageConfig(cfg))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)

	// Grab the DealID from the created deal.
	si1, err := fapi.StorageInfo(c)
	require.NoError(t, err)
	si1p := si1.Cold.Filecoin.Proposals[0]

	// Remove the Cid from the user.
	jid, err = fapi.PushStorageConfig(c, api.WithOverride(true), api.WithStorageConfig(cfg.WithColdEnabled(false)))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)
	err = fapi.Remove(c)
	require.NoError(t, err)
	ctx := context.Background()
	// Delete the cid data from go-ipfs, so we're clean.
	err = ipfsAPI.Dag().Remove(ctx, c)
	require.NoError(t, err)

	// Start a *new* storage config but with the deal import opt with the DealID
	// that still should be active in the Filecoin network.
	cfg = fapi.DefaultStorageConfig().WithHotEnabled(true).WithHotAllowUnfreeze(true)
	jid, err = fapi.PushStorageConfig(c, api.WithStorageConfig(cfg), api.WithDealImport([]uint64{si1p.DealID}))
	require.NoError(t, err)
	it.RequireEventualJobState(t, fapi, jid, ffs.Success)
	it.RequireStorageConfig(t, fapi, c, &cfg)

	// Check that the retrieved data matches the original data.
	re, err := fapi.Get(ctx, c)
	require.NoError(t, err)
	fetched, err := ioutil.ReadAll(re)
	require.NoError(t, err)
	require.True(t, bytes.Equal(data, fetched))
	it.RequireRetrievalDealRecord(t, fapi, c)
}
