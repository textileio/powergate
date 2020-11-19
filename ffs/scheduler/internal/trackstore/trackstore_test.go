package trackstore

import (
	"sort"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

var (
	scRepairable = ffs.StorageConfig{
		Repairable: true,
	}
	scRenewable = ffs.StorageConfig{
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				Renew: ffs.FilRenew{
					Enabled: true,
				},
			},
		},
	}
	scRepairableRenewable = ffs.StorageConfig{
		Repairable: true,
		Cold: ffs.ColdConfig{
			Enabled: true,
			Filecoin: ffs.FilConfig{
				Renew: ffs.FilRenew{
					Enabled: true,
				},
			},
		},
	}
)

func TestEmpty(t *testing.T) {
	t.Parallel()
	ts := create(t)

	repairables, err := ts.GetRepairables()
	require.NoError(t, err)
	require.Len(t, repairables, 0)

	renewables, err := ts.GetRenewables()
	require.NoError(t, err)
	require.Len(t, renewables, 0)
}

func TestPut(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	// Repairable
	c1, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c1, scRepairable)
	require.NoError(t, err)

	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid}})

	// Renewable
	c2, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs82")
	err = ts.Put(iid, c2, scRenewable)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c2, IIDs: []ffs.APIID{iid}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid}})
}

func TestDoublePut(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRepairable)
	require.NoError(t, err)
	err = ts.Put(iid, c, scRepairable)
	require.NoError(t, err)

	requireRepairables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})
}

func TestUninterestingPut(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	scUninteresting := ffs.StorageConfig{}
	err := ts.Put(iid, c, scUninteresting)
	require.NoError(t, err)

	requireRepairables(t, ts)
}

func TestRevertedPut(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	// Add some repariable
	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRepairable)
	require.NoError(t, err)
	requireRepairables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})

	// Put again but not repariable, should be gone.
	scUninteresting := ffs.StorageConfig{}
	err = ts.Put(iid, c, scUninteresting)
	require.NoError(t, err)
	requireRepairables(t, ts)
}

func TestPutRenewableRepairable(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRepairableRenewable)
	require.NoError(t, err)

	requireRepairables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})
	requireRenewables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})
}

func TestPutRenewableRepairableReverted(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRepairableRenewable)
	require.NoError(t, err)
	err = ts.Put(iid, c, scRenewable)
	require.NoError(t, err)
	requireRepairables(t, ts)
	requireRenewables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})
}

func TestPutRenewableSwitchRepairable(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRenewable)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})

	err = ts.Put(iid, c, scRepairable)
	require.NoError(t, err)
	requireRenewables(t, ts)
	requireRepairables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})
}

func TestRemove(t *testing.T) {
	t.Parallel()

	ts := create(t)
	iid := ffs.NewAPIID()

	c, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	err := ts.Put(iid, c, scRenewable)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c, IIDs: []ffs.APIID{iid}})

	err = ts.Remove(iid, c)
	require.NoError(t, err)

	requireRenewables(t, ts)
	requireRepairables(t, ts)

}

func TestPutMultipleIIDsWithRemoves(t *testing.T) {
	t.Parallel()

	ts := create(t)
	c1, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs81")
	c2, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs82")
	c3, _ := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs83")

	// IID1:
	// - c1 repairable
	// - c2 renewable
	iid1 := ffs.APIID("1")
	err := ts.Put(iid1, c1, scRepairable)
	require.NoError(t, err)
	err = ts.Put(iid1, c2, scRenewable)
	require.NoError(t, err)

	// IID2:
	// - c1 repairable
	// - c2 renewable
	iid2 := ffs.APIID("2")
	err = ts.Put(iid2, c1, scRepairable)
	require.NoError(t, err)
	err = ts.Put(iid2, c2, scRenewable)
	require.NoError(t, err)

	requireRenewables(t, ts, expected{Cid: c2, IIDs: []ffs.APIID{iid1, iid2}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid1, iid2}})

	// IID3:
	// - c1 renewable
	// - c3 renewable
	iid3 := ffs.APIID("3")
	err = ts.Put(iid3, c1, scRenewable)
	require.NoError(t, err)
	err = ts.Put(iid3, c3, scRenewable)
	require.NoError(t, err)

	requireRenewables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid3}}, expected{Cid: c2, IIDs: []ffs.APIID{iid1, iid2}}, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid1, iid2}})

	// Remove arbitrarly and check correctness.
	err = ts.Remove(iid2, c2)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid3}}, expected{Cid: c2, IIDs: []ffs.APIID{iid1}}, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid1, iid2}})

	err = ts.Remove(iid3, c1)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c2, IIDs: []ffs.APIID{iid1}}, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid1, iid2}})

	err = ts.Remove(iid1, c1)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c2, IIDs: []ffs.APIID{iid1}}, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid2}})

	err = ts.Remove(iid1, c2)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts, expected{Cid: c1, IIDs: []ffs.APIID{iid2}})

	err = ts.Remove(iid2, c1)
	require.NoError(t, err)
	requireRenewables(t, ts, expected{Cid: c3, IIDs: []ffs.APIID{iid3}})
	requireRepairables(t, ts)

	err = ts.Remove(iid3, c3)
	require.NoError(t, err)
	requireRenewables(t, ts)
	requireRepairables(t, ts)
}

type expected struct {
	Cid  cid.Cid
	IIDs []ffs.APIID
}

func requireRepairables(t *testing.T, ts *Store, exp ...expected) {
	t.Helper()
	tcs, err := ts.GetRepairables()
	require.NoError(t, err)
	require.Len(t, tcs, len(exp))

	sortTC(tcs)
	for i := range exp {
		require.Equal(t, exp[i].Cid, tcs[i].Cid)
		for j := range exp[i].IIDs {
			require.Equal(t, exp[i].IIDs[j], tcs[i].Tracked[j].IID)
			require.True(t, tcs[i].Tracked[j].StorageConfig.Repairable)
		}
	}
}

func requireRenewables(t *testing.T, ts *Store, exp ...expected) {
	t.Helper()
	tcs, err := ts.GetRenewables()
	require.NoError(t, err)
	require.Len(t, tcs, len(exp))

	sortTC(tcs)
	for i := range exp {
		require.Equal(t, exp[i].Cid, tcs[i].Cid)
		for j := range exp[i].IIDs {
			require.Equal(t, exp[i].IIDs[j], tcs[i].Tracked[j].IID)
			sc := tcs[i].Tracked[j].StorageConfig
			require.True(t, sc.Cold.Enabled)
			require.True(t, sc.Cold.Filecoin.Renew.Enabled)
		}
	}
}

func sortTC(tcs []TrackedCid) {
	sort.Slice(tcs, func(i, j int) bool {
		return tcs[i].Cid.String() < tcs[j].Cid.String()
	})
	for _, tc := range tcs {
		sort.Slice(tc.Tracked, func(i, j int) bool {
			return tc.Tracked[i].IID.String() < tc.Tracked[j].IID.String()
		})
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds)
	require.NoError(t, err)
	return store
}
