package trackstore

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestAll(t *testing.T) {
	ts := create(t)

	// Repairable
	c1, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	_, _, err := ts.Get(c1)
	require.Error(t, err)

	iid1 := ffs.NewAPIID()
	scRepairable := ffs.StorageConfig{
		Repairable: true,
	}
	err = ts.Put(iid1, c1, scRepairable)
	require.NoError(t, err)

	// Renewable
	c2, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8r")
	iid2 := ffs.NewAPIID()
	scRenewable := ffs.StorageConfig{
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				Renew: ffs.FilRenew{
					Enabled: true,
				},
			},
		},
	}
	err = ts.Put(iid2, c2, scRenewable)
	require.NoError(t, err)

	scT, iidT, err := ts.Get(c1)
	require.NoError(t, err)
	require.Equal(t, scRepairable, scT)
	require.Equal(t, iid1, iidT)
	scT, iidT, err = ts.Get(c2)
	require.NoError(t, err)
	require.Equal(t, scRenewable, scT)
	require.Equal(t, iid2, iidT)

	// Check repairables
	repairables, err := ts.GetRepairables()
	require.NoError(t, err)
	require.Len(t, repairables, 1)
	require.Equal(t, c1, repairables[0])

	// Check renewables
	renewables, err := ts.GetRenewables()
	require.NoError(t, err)
	require.Len(t, renewables, 1)
	require.Equal(t, c2, renewables[0])

	// Remove and check correct removals
	err = ts.Remove(c1)
	require.NoError(t, err)
	_, _, err = ts.Get(c1)
	require.Error(t, err)
	repairables, err = ts.GetRepairables()
	require.NoError(t, err)
	require.Len(t, repairables, 0)

	// Test if loading caches works correctly.
	err = ts.loadCaches()
	require.NoError(t, err)
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds)
	require.NoError(t, err)
	return store
}
