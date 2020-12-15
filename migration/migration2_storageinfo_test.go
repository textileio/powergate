package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

func TestV2_StorageInfo(t *testing.T) {
	t.Parallel()

	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v2_StorageInfo.pre")

	propDealID, err := v2GenProposalCidToDealID(ds)
	require.NoError(t, err)

	txn, _ := ds.NewTransaction(false)
	err = v2MigrationDealIDFilling(txn, propDealID)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v2_StorageInfo.post")
}
