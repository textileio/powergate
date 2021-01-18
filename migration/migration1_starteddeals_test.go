package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/tests"
	"github.com/textileio/powergate/v2/util"
)

func TestV1_StartedDeals(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_StartedDeals.pre")

	txn, _ := ds.NewTransaction(false)
	err := migrateStartedDeals(txn)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v1_StartedDeals.post")
}

func TestV0_ExecutingJobs(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_StartedDeals.pre")

	txn, _ := ds.NewTransaction(false)
	execJobsOwners, err := v0GetExecutingJobs(txn)
	require.NoError(t, err)

	require.Len(t, execJobsOwners, 2)

	c1, _ := util.CidFromString("QmbcKgdxWfZzePsrc6rWq2yk12GE1MnW8JJiubHAJovKb5")
	iid1 := ffs.APIID("a9aeaecc-4c94-4bb6-bd14-13927d8cede0")
	require.Equal(t, iid1, execJobsOwners[c1])

	c2, _ := util.CidFromString("QmaPndBy99qcz7rCAooak8SY34Rk9oXWpEBp4dUyhbmq1q")
	iid2 := ffs.APIID("ad2f3b0c-e356-43d4-a483-fba79479d7e4")
	require.Equal(t, iid2, execJobsOwners[c2])
}
