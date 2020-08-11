package chainstore

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/google/go-cmp/cmp"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

type data struct {
	Tipset string
	Nested extraData
}

type extraData struct {
	Pos int
}

func TestLoadFromEmpty(t *testing.T) {
	ctx := context.Background()
	cs, err := New(tests.NewTxMapDatastore(), newMockTipsetOrderer())
	require.NoError(t, err)

	var d data
	target := types.NewTipSetKey(cid.Undef)
	ts, err := cs.LoadAndPrune(ctx, target, &d)
	require.NoError(t, err)

	require.Nil(t, ts, "base tipset should be nil")

	require.Equal(t, data{}, d, "state should be the default value")
}

func TestSaveSingle(t *testing.T) {
	ctx := context.Background()
	mto := newMockTipsetOrderer()

	cs, err := New(tests.NewTxMapDatastore(), mto)
	require.NoError(t, err)

	ts, v := mto.next(t)
	err = cs.Save(ctx, ts, &v)
	require.NoError(t, err)

	var v2 data
	bts, err := cs.LoadAndPrune(ctx, ts, &v2)
	require.NoError(t, err)

	if !cmp.Equal(v, v2) || *bts != ts {
		t.Fatalf("saved and loaded state from same tipset should be equal")
	}
}

func TestSaveMultiple(t *testing.T) {
	ctx := context.Background()
	mto := newMockTipsetOrderer()

	cs, err := New(tests.NewTxMapDatastore(), mto)
	require.NoError(t, err)

	generateTotal := 100
	for i := 0; i < generateTotal; i++ {
		ts, v := mto.next(t)
		err := cs.Save(ctx, ts, &v)
		require.NoError(t, err)
	}

	// Check that we're capping # of checkpoints to maxCheckpoints
	if len(cs.checkpoints) != maxCheckpoints {
		t.Fatalf("there should be exactly maxCheckpoints saved")
	}
	// Check saved ones are the last maxCheckpoint ones
	expectedTipsets := mto.list[generateTotal-maxCheckpoints:]
	for i, c := range cs.checkpoints {
		require.Equal(t, expectedTipsets[i], c.ts, "saved tipset doesn't correspond with expected one")
	}

	for i := len(expectedTipsets) - 1; i >= 0; i-- {
		ts := expectedTipsets[i]
		var v data
		bts, err := cs.LoadAndPrune(ctx, ts, &v)
		require.NoError(t, err)
		if *bts != ts || v.Nested.Pos != generateTotal-maxCheckpoints+i {
			t.Fatalf("elem %d doesn't seem to be loaded from correct tipset", i)
		}
	}
}

func TestSaveInvalid(t *testing.T) {
	ctx := context.Background()
	mto := newMockTipsetOrderer()
	cs, err := New(tests.NewTxMapDatastore(), mto)
	require.NoError(t, err)

	ts1, v1 := mto.next(t)
	ts2, v2 := mto.next(t)

	err = cs.Save(ctx, ts2, &v2)
	require.NoError(t, err)

	err = cs.Save(ctx, ts1, &v1)
	require.Error(t, err, "Save must not allow to save state on an older tipset that last known")
}

// Most interesting test.
// Create 10 happy-chain tipset saves. Load from new tipset (tsFork) that forks from
// 6th saved tipset. The Load should delete checkpoints 7, 8, 9 and 10 since tsFork
// doesnt Precede() from any of them, and return state of checkpoint 6.
// Saying it differently, Load should return the last state from the most recent
// checkpoint that precedes the target tipset.
func TestLoadForkedCheckpoint(t *testing.T) {
	ctx := context.Background()
	mto := newMockTipsetOrderer()

	cs, err := New(tests.NewTxMapDatastore(), mto)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		ts, v := mto.next(t)
		err := cs.Save(ctx, ts, &v)
		require.NoError(t, err)
	}

	fts := mto.fork(t, 5)
	var v data
	bts, err := cs.LoadAndPrune(ctx, fts, &v)
	require.NoError(t, err)

	if *bts != mto.list[5] {
		t.Fatalf("returned base tipset state should be from the 6th checkpoint")
	}
	if v.Nested.Pos != 5 {
		t.Fatalf("state return doesn't seem to correspond to the 6th checkpoint")
	}
}

func TestLoadSavedState(t *testing.T) {
	ctx := context.Background()
	mto := newMockTipsetOrderer()
	ds := tests.NewTxMapDatastore()
	cs, err := New(ds, mto)
	require.NoError(t, err)

	generateTotal := 100
	for i := 0; i < generateTotal; i++ {
		ts, v := mto.next(t)
		err := cs.Save(ctx, ts, &v)
		require.NoError(t, err)
	}

	cs, err = New(ds, mto)
	require.NoError(t, err)
	if len(cs.checkpoints) != maxCheckpoints {
		t.Fatalf("checkpoints are missing")
	}

	offset := 3
	savedTipset := mto.list[len(mto.list)-offset]
	var v data
	bts, err := cs.LoadAndPrune(ctx, savedTipset, &v)
	require.NoError(t, err)
	if *bts != savedTipset || v.Tipset != savedTipset.String() || v.Nested.Pos != generateTotal-offset {
		t.Fatalf("returned state is wrong")
	}
}

type mockTipsetOrderer struct {
	forks map[string]string
	list  []types.TipSetKey
}

func newMockTipsetOrderer() *mockTipsetOrderer {
	return &mockTipsetOrderer{
		forks: make(map[string]string),
	}
}

func (mto *mockTipsetOrderer) Precedes(ctx context.Context, from, to types.TipSetKey) (bool, error) {
	if forkedTs, ok := mto.forks[from.String()]; ok {
		if forkedTs == to.String() {
			return true, nil
		}
	}

	var foundFrom bool
	for _, v := range mto.list {
		foundFrom = foundFrom || from == v
		if foundFrom && to == v {
			return true, nil
		}
	}
	return false, nil
}

func (mto *mockTipsetOrderer) next(t *testing.T) (types.TipSetKey, data) {
	ts := randomTipsetkey(t)
	mto.list = append(mto.list, ts)

	return ts, data{Tipset: ts.String(), Nested: extraData{
		Pos: len(mto.list) - 1,
	}}
}

func (mto *mockTipsetOrderer) fork(t *testing.T, i int) types.TipSetKey {
	fork := randomTipsetkey(t)
	mto.forks[mto.list[i].String()] = fork.String()
	return fork
}

func randomTipsetkey(t *testing.T) types.TipSetKey {
	r := make([]byte, 16)
	_, err := rand.Read(r)
	require.NoError(t, err)

	mh, err := multihash.Sum(r, multihash.IDENTITY, -1)
	require.NoError(t, err)
	return types.NewTipSetKey(cid.NewCidV1(cid.Raw, mh))
}
