package migration

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestV1_StorageInfo(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_StorageInfo.pre")

	c1, _ := util.CidFromString("QmbtfAvVRVgEa9RH6vYpKg1HWbBQjxiZ6vNG1AAvt3vyfR")
	c2, _ := util.CidFromString("QmZTMaDfCMWqhUXYDnKup8ctCTyPxnriYW7G4JR8KXoX5M")
	cidOwners := map[cid.Cid][]ffs.APIID{
		c1: {ffs.APIID("ID1"), ffs.APIID("ID2")},
		c2: {ffs.APIID("ID2"), ffs.APIID("ID3")},
	}
	txn, _ := ds.NewTransaction(false)
	err := migrateStorageInfo(txn, cidOwners)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v1_StorageInfo.post")
}
