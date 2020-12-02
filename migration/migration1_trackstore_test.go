package migration

import (
	"sort"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestV1_Trackstore(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_Trackstore.pre")

	c1, _ := util.CidFromString("QmY7gN6AfKSoR7DNEjcUyXRYS85giD1YXN62cWVzS5zfus")
	c2, _ := util.CidFromString("QmX5J6NujFycQyoMvHXjTNmtvqDnn8TN6wAVQJ4Ap2GMnq")
	cidOwners := map[cid.Cid][]ffs.APIID{
		c1: {ffs.APIID("ad2f3b0c-e356-43d4-a483-fba79479d7e4"), ffs.APIID("fb79f525-a3c4-47f0-94e5-e344c5b1dec5")},
		c2: {ffs.APIID("2bef4790-a47a-4a48-90da-a89f93ab6310")},
	}
	txn, _ := ds.NewTransaction(false)
	err := migrateTrackstore(txn, cidOwners)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v1_Trackstore.post")
}

func TestV0_APIIDs(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_Trackstore.pre")

	txn, _ := ds.NewTransaction(false)
	iids, err := v0APIIDs(txn)
	require.NoError(t, err)

	sort.Slice(iids, func(i, j int) bool {
		return iids[i].String() < iids[j].String()
	})

	require.Len(t, iids, 3)
	require.Equal(t, "2bef4790-a47a-4a48-90da-a89f93ab6310", iids[0].String())
	require.Equal(t, "ad2f3b0c-e356-43d4-a483-fba79479d7e4", iids[1].String())
	require.Equal(t, "fb79f525-a3c4-47f0-94e5-e344c5b1dec5", iids[2].String())
}

func TestV0_GetCidsFromIID(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_Trackstore.pre")

	txn, _ := ds.NewTransaction(false)
	cids, err := v0GetCidsFromIID(txn, "ad2f3b0c-e356-43d4-a483-fba79479d7e4")
	require.NoError(t, err)

	require.Len(t, cids, 1)
	require.Equal(t, "QmY7gN6AfKSoR7DNEjcUyXRYS85giD1YXN62cWVzS5zfus", cids[0].String())
}

func TestV0_CidOwners(t *testing.T) {
	t.Parallel()
	c1, _ := util.CidFromString("QmY7gN6AfKSoR7DNEjcUyXRYS85giD1YXN62cWVzS5zfus")
	c2, _ := util.CidFromString("QmX5J6NujFycQyoMvHXjTNmtvqDnn8TN6wAVQJ4Ap2GMnq")
	expectedOwners := map[cid.Cid][]ffs.APIID{
		c1: {ffs.APIID("ad2f3b0c-e356-43d4-a483-fba79479d7e4"), ffs.APIID("fb79f525-a3c4-47f0-94e5-e344c5b1dec5")},
		c2: {ffs.APIID("2bef4790-a47a-4a48-90da-a89f93ab6310")},
	}
	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_Trackstore.pre")

	txn, _ := ds.NewTransaction(false)
	owners, err := v0CidOwners(txn)
	require.NoError(t, err)

	require.Equal(t, len(expectedOwners), len(owners))
	for k1, v1 := range expectedOwners {
		v2, ok := owners[k1]
		require.True(t, ok)
		sort.Slice(v2, func(i, j int) bool {
			return v2[i].String() < v2[j].String()
		})
		require.Equal(t, v1, v2)
	}
}
