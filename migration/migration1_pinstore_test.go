package migration

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestV1_Pinstore(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_Pinstore.pre")

	c1, _ := util.CidFromString("QmY7gN6AfKSoR7DNEjcUyXRYS85giD1YXN62cWVzS5zfus")
	c2, _ := util.CidFromString("QmX5J6NujFycQyoMvHXjTNmtvqDnn8TN6wAVQJ4Ap2GMnq")
	cidOwners := map[cid.Cid][]ffs.APIID{
		c1: {ffs.APIID("ad2f3b0c-e356-43d4-a483-fba79479d7e4"), ffs.APIID("fb79f525-a3c4-47f0-94e5-e344c5b1dec5")},
		c2: {ffs.APIID("2bef4790-a47a-4a48-90da-a89f93ab6310")},
	}
	txn, _ := ds.NewTransaction(false)
	err := pinstoreFilling(txn, cidOwners)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v1_Pinstore.post")
}
