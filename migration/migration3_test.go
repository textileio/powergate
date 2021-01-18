package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/tests"
)

func TestV3(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v3_StorageJobs.pre")
	txn, _ := ds.NewTransaction(false)

	err := V3StorageJobsIndexMigration.Run(txn)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v3_StorageJobs.post")
}
