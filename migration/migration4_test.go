package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/tests"
)

func TestV4(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v4_Records.pre")
	txn, _ := ds.NewTransaction(false)

	err := V4RecordsMigration.Run(txn)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v4_Records.post")
}
