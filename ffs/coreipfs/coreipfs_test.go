package coreipfs

import (
	"bytes"
	"context"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	it "github.com/textileio/powergate/ffs/integrationtest"
	"github.com/textileio/powergate/ffs/joblogger"
	"github.com/textileio/powergate/tests"
	txndstr "github.com/textileio/powergate/txndstransform"
)

func TestStage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)
	iid := ffs.NewAPIID()

	c, err := ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsGCable(t, ci, c)
	okPinned, err := ci.IsPinned(ctx, iid, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 0, 1)

	// Re-stage and test ref count is stil 1.
	c, err = ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	requireRefCount(t, ci, c, 0, 1)
}

func TestStagePinStage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)
	iid := ffs.NewAPIID()

	// Stage
	c, err := ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	requireRefCount(t, ci, c, 0, 1)

	// Pin that Cid.
	_, err = ci.Pin(ctx, iid, c)
	require.NoError(t, err)
	requireRefCount(t, ci, c, 1, 0)

	// Stage again, it shouldn't be GCable.
	c2, err := ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, c, c2)
	it.RequireIpfsPinnedCid(ctx, t, c2, ipfs)
	requireCidIsNotGCable(t, ci, c2)
	okPinned, err := ci.IsPinned(ctx, iid, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 1, 0)
}

func TestPinAndRePin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)
	iid := ffs.NewAPIID()

	rd := ipfsfiles.NewReaderFile(bytes.NewReader(data))
	p, err := ipfs.Unixfs().Add(ctx, rd, options.Unixfs.Pin(false))
	require.NoError(t, err)
	c := p.Cid()

	size, err := ci.Pin(ctx, iid, c)
	require.NoError(t, err)
	require.Greater(t, size, 0)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsNotGCable(t, ci, c)
	okPinned, err := ci.IsPinned(ctx, iid, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 1, 0)

	// Pin again, check that ref count is still 1.
	_, err = ci.Pin(ctx, iid, c)
	require.NoError(t, err)
	requireRefCount(t, ci, c, 1, 0)
}

func TestUnpin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)
	iid := ffs.NewAPIID()

	rd := ipfsfiles.NewReaderFile(bytes.NewReader(data))
	p, err := ipfs.Unixfs().Add(ctx, rd, options.Unixfs.Pin(false))
	require.NoError(t, err)
	c := p.Cid()

	// Test unpin an unpinned cid.
	err = ci.Unpin(ctx, iid, c)
	require.Equal(t, ErrUnpinnedCid, err)

	// Test pin and unpin.
	requireRefCount(t, ci, c, 0, 0)
	_, err = ci.Pin(ctx, iid, c)
	require.NoError(t, err)

	err = ci.Unpin(ctx, iid, c)
	require.NoError(t, err)
	requireRefCount(t, ci, c, 0, 0)

	// Test unpin an unpinned cid again.
	err = ci.Unpin(ctx, iid, c)
	require.Equal(t, ErrUnpinnedCid, err)
}

func TestReplaceThatUnpinAndPin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	iid := ffs.NewAPIID()

	// Pin c1
	data := it.RandomBytes(r, 1500)
	c1, err := ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	_, err = ci.Pin(ctx, iid, c1)
	require.NoError(t, err)
	requireRefCount(t, ci, c1, 1, 0)

	// Stage c2
	data2 := it.RandomBytes(r, 1500)
	c2, err := ci.Stage(ctx, iid, bytes.NewReader(data2))
	require.NoError(t, err)
	requireRefCount(t, ci, c2, 0, 1)

	// Replace
	size, err := ci.Replace(ctx, iid, c1, c2)
	require.NoError(t, err)
	require.Greater(t, size, 0)

	// Post checks
	it.RequireIpfsUnpinnedCid(ctx, t, c1, ipfs) // c1 unpinned in node.
	it.RequireIpfsPinnedCid(ctx, t, c2, ipfs)   // c2 pinned in node.

	okPinned, err := ci.IsPinned(ctx, iid, c1)
	require.NoError(t, err)
	require.False(t, okPinned) // API claims c1 is unpinned.
	okPinned, err = ci.IsPinned(ctx, iid, c2)
	require.NoError(t, err)
	require.True(t, okPinned) // API claims c2 is unpinned.

	requireRefCount(t, ci, c1, 0, 0) // c1 ref count all 0.
	requireRefCount(t, ci, c2, 1, 0) // c2 from staged to strong.
}

func TestReplaceNotUnpinAndPin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)

	// Pin c1
	iid1 := ffs.NewAPIID()
	data := it.RandomBytes(r, 1500)
	c1, err := ci.Stage(ctx, iid1, bytes.NewReader(data))
	require.NoError(t, err)
	_, err = ci.Pin(ctx, iid1, c1)
	require.NoError(t, err)
	requireRefCount(t, ci, c1, 1, 0)

	// Make another iid pin c1, so can't be unpinned by replace.
	iid2 := ffs.NewAPIID()
	_, err = ci.Pin(ctx, iid2, c1)
	require.NoError(t, err)
	requireRefCount(t, ci, c1, 2, 0)

	// Stage data2
	data2 := it.RandomBytes(r, 1500)
	c2, err := ci.Stage(ctx, iid1, bytes.NewReader(data2))
	require.NoError(t, err)
	requireRefCount(t, ci, c2, 0, 1)

	// Replace
	size, err := ci.Replace(ctx, iid1, c1, c2)
	require.NoError(t, err)
	require.Greater(t, size, 0)

	// Post checks
	it.RequireIpfsPinnedCid(ctx, t, c1, ipfs) // c1 still pinned in node, since used by iid2.
	it.RequireIpfsPinnedCid(ctx, t, c2, ipfs) // c2 pinned in node.

	okPinned, err := ci.IsPinned(ctx, iid1, c1)
	require.NoError(t, err)
	require.False(t, okPinned) // API claims c1 unpinned in iid1.
	okPinned, err = ci.IsPinned(ctx, iid2, c1)
	require.NoError(t, err)
	require.True(t, okPinned) // API claims c1 pinned in iid2.
	okPinned, err = ci.IsPinned(ctx, iid1, c2)
	require.NoError(t, err)
	require.True(t, okPinned) // API claims c2 pinned in iid1.

	requireRefCount(t, ci, c1, 1, 0) // from (2, 0) to (1, 0), only iid2.
	requireRefCount(t, ci, c2, 1, 0) // c2 strong pin by replace.
}

func TestReplaceErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	iid := ffs.NewAPIID()

	// Add directly to IPFS node without pinning
	data := it.RandomBytes(r, 1500)
	rd := ipfsfiles.NewReaderFile(bytes.NewReader(data))
	p, err := ipfs.Unixfs().Add(ctx, rd, options.Unixfs.Pin(false))
	require.NoError(t, err)
	c1 := p.Cid()

	// Stage c2
	data2 := it.RandomBytes(r, 1500)
	c2, err := ci.Stage(ctx, iid, bytes.NewReader(data2))
	require.NoError(t, err)
	requireRefCount(t, ci, c2, 0, 1)

	// Replace
	_, err = ci.Replace(ctx, iid, c1, c2)
	require.Equal(t, ErrReplaceFromNotPinned, err)
	requireRefCount(t, ci, c1, 0, 0)
	requireRefCount(t, ci, c2, 0, 1)
}

// Test pinning a Cid, unpinning it, and Stage it again.
func TestPinUnpinStage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)
	iid := ffs.NewAPIID()

	rd := ipfsfiles.NewReaderFile(bytes.NewReader(data))
	p, err := ipfs.Unixfs().Add(ctx, rd, options.Unixfs.Pin(false))
	require.NoError(t, err)
	c := p.Cid()

	_, err = ci.Pin(ctx, iid, c)
	require.NoError(t, err)
	err = ci.Unpin(ctx, iid, c)
	require.NoError(t, err)
	requireRefCount(t, ci, c, 0, 0)

	// Stage it again and check invariants:
	// must be gcable, is pinned, and should have a staged-pin.
	c, err = ci.Stage(ctx, iid, bytes.NewReader(data))
	require.NoError(t, err)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsGCable(t, ci, c)
	okPinned, err := ci.IsPinned(ctx, iid, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 0, 1)
}

func TestMultipleStage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, ipfs := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)

	// Stage with iid1.
	iid1 := ffs.NewAPIID()
	c, err := ci.Stage(ctx, iid1, bytes.NewReader(data))
	require.NoError(t, err)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsGCable(t, ci, c)
	okPinned, err := ci.IsPinned(ctx, iid1, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 0, 1)

	// Stage with iid1.
	iid2 := ffs.NewAPIID()
	c, err = ci.Stage(ctx, iid2, bytes.NewReader(data))
	require.NoError(t, err)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsGCable(t, ci, c)
	okPinned, err = ci.IsPinned(ctx, iid2, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 0, 2)

	// Stage with iid3.
	iid3 := ffs.NewAPIID()
	c, err = ci.Stage(ctx, iid3, bytes.NewReader(data))
	require.NoError(t, err)
	it.RequireIpfsPinnedCid(ctx, t, c, ipfs)
	requireCidIsGCable(t, ci, c)
	okPinned, err = ci.IsPinned(ctx, iid3, c)
	require.NoError(t, err)
	require.True(t, okPinned)
	requireRefCount(t, ci, c, 0, 3)
}

func TestTwoStageOnePin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, _ := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)

	// Stage with iid1
	iid1 := ffs.NewAPIID()
	c, err := ci.Stage(ctx, iid1, bytes.NewReader(data))
	require.NoError(t, err)
	require.True(t, c.Defined())

	// Stage with other iid.
	iid2 := ffs.NewAPIID()
	c, err = ci.Stage(ctx, iid2, bytes.NewReader(data))
	require.NoError(t, err)

	requireCidIsGCable(t, ci, c) // Can be GCable

	// Pin with iid1
	_, err = ci.Pin(ctx, iid1, c)
	require.NoError(t, err)
	requireCidIsNotGCable(t, ci, c) // Can't be GCable
	requireRefCount(t, ci, c, 1, 1) // One strong and one staged.

	// Now unpin and check.
	err = ci.Unpin(ctx, iid1, c)
	require.NoError(t, err)
	requireCidIsGCable(t, ci, c)    // Now is GCable again.
	requireRefCount(t, ci, c, 0, 1) // Only iid2 staged.
}

func TestPinnedCids(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := rand.New(rand.NewSource(22))

	ci, _ := newCoreIPFS(t)
	data := it.RandomBytes(r, 1500)

	// Stage with iid1
	iid1 := ffs.NewAPIID()
	c1, err := ci.Stage(ctx, iid1, bytes.NewReader(data))
	require.NoError(t, err)

	// Stage with iid2.
	iid2 := ffs.NewAPIID()
	_, err = ci.Stage(ctx, iid2, bytes.NewReader(data))
	require.NoError(t, err)

	// Pin with iid1
	_, err = ci.Pin(ctx, iid1, c1)
	require.NoError(t, err)

	// Stage another cid with iid3.
	data = it.RandomBytes(r, 1500)
	iid3 := ffs.NewAPIID()
	c2, err := ci.Stage(ctx, iid3, bytes.NewReader(data))
	require.NoError(t, err)

	// Get all and check.
	all, err := ci.PinnedCids(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)

	// Order the slices so we have predictable results
	// for comparing. In position 0 is c2, and 1 is c1.
	sort.Slice(all, func(a, b int) bool {
		return all[a].Cid.String() < all[b].Cid.String()
	})

	// c2 is staged by iid3.
	require.Equal(t, c2, all[0].Cid)
	require.Len(t, all[0].APIIDs, 1)
	require.Equal(t, iid3, all[0].APIIDs[0].ID)
	require.True(t, all[0].APIIDs[0].Staged)

	// c1 is:
	// - pinned by iid1
	// - staged by iid2
	require.Equal(t, c1, all[1].Cid)
	require.Len(t, all[1].APIIDs, 2)
	require.Equal(t, iid1, all[1].APIIDs[0].ID)
	require.False(t, all[1].APIIDs[0].Staged)
	require.Equal(t, iid2, all[1].APIIDs[1].ID)
	require.True(t, all[1].APIIDs[1].Staged)
}

func TestGCSingleAPIID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	iid := ffs.NewAPIID()

	t.Run("Simple", func(t *testing.T) {
		ci, ipfs := newCoreIPFS(t)
		// # Stage 1
		r := rand.New(rand.NewSource(22))
		data := it.RandomBytes(r, 1500)
		c1, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		// # Stage 1
		data = it.RandomBytes(r, 1500)
		c2, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		gced, err := ci.GCStaged(ctx, nil, time.Now())
		require.NoError(t, err)
		require.Len(t, gced, 2)

		it.RequireIpfsUnpinnedCid(ctx, t, c1, ipfs)
		it.RequireIpfsUnpinnedCid(ctx, t, c2, ipfs)

		gced, err = ci.GCStaged(ctx, nil, time.Now())
		require.NoError(t, err)
		require.Len(t, gced, 0)
	})

	t.Run("Exclusion", func(t *testing.T) {
		ci, ipfs := newCoreIPFS(t)
		// # Stage 1
		r := rand.New(rand.NewSource(22))
		data := it.RandomBytes(r, 1500)
		c1, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		// # Stage 1
		data = it.RandomBytes(r, 1500)
		c2, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		gced, err := ci.GCStaged(ctx, []cid.Cid{c1}, time.Now())
		require.NoError(t, err)
		require.Len(t, gced, 1)

		it.RequireIpfsUnpinnedCid(ctx, t, c2, ipfs)

		gced, err = ci.GCStaged(ctx, nil, time.Now())
		require.NoError(t, err)
		require.Len(t, gced, 1)
		it.RequireIpfsUnpinnedCid(ctx, t, c1, ipfs)
	})

	t.Run("Very old", func(t *testing.T) {
		ci, ipfs := newCoreIPFS(t)
		// # Stage 1
		r := rand.New(rand.NewSource(22))
		data := it.RandomBytes(r, 1500)
		c1, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		// # Stage 1
		data = it.RandomBytes(r, 1500)
		c2, err := ci.Stage(ctx, iid, bytes.NewReader(data))
		require.NoError(t, err)

		gced, err := ci.GCStaged(ctx, nil, time.Now().Add(-time.Hour))
		require.NoError(t, err)
		require.Len(t, gced, 0)

		gced, err = ci.GCStaged(ctx, nil, time.Now())
		require.NoError(t, err)
		require.Len(t, gced, 2)

		it.RequireIpfsUnpinnedCid(ctx, t, c1, ipfs)
		it.RequireIpfsUnpinnedCid(ctx, t, c2, ipfs)
	})
}

func requireCidIsGCable(t *testing.T, ci *CoreIpfs, c cid.Cid) {
	t.Helper()
	require.True(t, isGCable(t, ci, c))
}

func requireCidIsNotGCable(t *testing.T, ci *CoreIpfs, c cid.Cid) {
	t.Helper()
	require.False(t, isGCable(t, ci, c))
}

func isGCable(t *testing.T, ci *CoreIpfs, c cid.Cid) bool {
	lst, err := ci.getGCCandidates(nil, time.Now())
	require.NoError(t, err)

	for _, cid := range lst {
		if cid.Equals(c) {
			return true
		}
	}
	return false
}

func requireRefCount(t *testing.T, ci *CoreIpfs, c cid.Cid, reqStrong, reqStaged int) {
	t.Helper()
	total, staged := ci.ps.RefCount(c)
	strong := total - staged

	require.Equal(t, strong, reqStrong)
	require.Equal(t, staged, reqStaged)
}

func newCoreIPFS(t *testing.T) (*CoreIpfs, *httpapi.HttpApi) {
	ds := tests.NewTxMapDatastore()
	ipfs, _ := it.CreateIPFS(t)
	l := joblogger.New(txndstr.Wrap(ds, "ffs/joblogger"))
	hl, err := New(ds, ipfs, l)
	require.NoError(t, err)

	return hl, ipfs
}
