package sjstore

import (
	"errors"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestEnqueue(t *testing.T) {
	t.Parallel()
	s := create(t)
	j := createJob(t)
	err := s.Enqueue(j)
	require.NoError(t, err)
	jQueued, err := s.Get(j.ID)
	require.NoError(t, err)
	j.Status = ffs.Queued
	require.Equal(t, j, jQueued)
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob(t)
		err := s.Enqueue(j)
		require.NoError(t, err)
		j2, err := s.Dequeue()
		require.NoError(t, err)
		j.Status = ffs.Executing
		require.NotNil(t, j2)
		require.Equal(t, j, *j2)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j, err := s.Dequeue()
		require.NoError(t, err)
		require.Nil(t, j)
	})
	t.Run("ExecutingAndFinalized", func(t *testing.T) {
		t.Parallel()
		s := create(t)

		j1 := createJob(t)
		err := s.Enqueue(j1)
		require.NoError(t, err)
		_, err = s.Dequeue()
		require.NoError(t, err)

		j2 := createJob(t)
		err = s.Enqueue(j2)
		require.NoError(t, err)

		// Dequeue should be nil since j1 is Executing,
		// and j2 can't be returned until that fininishes since
		// both are jobs for the same Cid.
		jq, err := s.Dequeue()
		require.NoError(t, err)
		require.Nil(t, jq)

		errors := []ffs.DealError{{ProposalCid: cid.Undef, Miner: "t0100", Message: "abcd"}}
		err = s.Finalize(j1.ID, ffs.Success, nil, errors)
		require.NoError(t, err)
		// Dequeue now returns a new job since the Executing one
		// was finalized
		jq, err = s.Dequeue()
		require.NoError(t, err)
		require.NotNil(t, jq)
		j2.Status = ffs.Executing
		require.Equal(t, j2, *jq)
	})
}

func TestCancelation(t *testing.T) {
	t.Parallel()
	s := create(t)

	j1 := createJob(t)
	err := s.Enqueue(j1)
	require.NoError(t, err)

	j2 := createJob(t)
	err = s.Enqueue(j2)
	require.NoError(t, err)

	j1Canceled, err := s.Get(j1.ID)
	require.NoError(t, err)
	require.Equal(t, ffs.Canceled, j1Canceled.Status)

	j2Queued, err := s.Get(j2.ID)
	require.NoError(t, err)
	require.Equal(t, ffs.Queued, j2Queued.Status)
}

func TestStartedDeals(t *testing.T) {
	t.Parallel()
	s := create(t)

	iid := ffs.NewAPIID()
	b, _ := multihash.Encode([]byte("data"), multihash.SHA1)
	cidData := cid.NewCidV1(1, b)

	fds, err := s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, 0, len(fds))

	b, _ = multihash.Encode([]byte("prop1"), multihash.SHA1)
	cidProp1 := cid.NewCidV1(1, b)
	b, _ = multihash.Encode([]byte("prop2"), multihash.SHA1)
	cidProp2 := cid.NewCidV1(1, b)

	startedDeals := []cid.Cid{cidProp1, cidProp2}
	err = s.AddStartedDeals(iid, cidData, startedDeals)
	require.NoError(t, err)

	fds, err = s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, startedDeals, fds)

	err = s.RemoveStartedDeals(iid, cidData)
	require.NoError(t, err)

	fds, err = s.GetStartedDeals(iid, cidData)
	require.NoError(t, err)
	require.Equal(t, 0, len(fds))
}

func TestQueryJobs(t *testing.T) {
	t.Run("ExecutingAndFailed", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob(t)

		err := s.Enqueue(j)
		require.NoError(t, err)

		queryJobs(t, j, s.QueuedJobs, 1)
		queryJobs(t, j, s.ExecutingJobs, 0)
		queryJobs(t, j, s.LatestFinalJobs, 0)
		queryJobs(t, j, s.LatestSuccessfulJobs, 0)

		dequeued, err := s.Dequeue()
		require.NoError(t, err)
		require.Equal(t, dequeued.ID, j.ID)

		queryJobs(t, j, s.QueuedJobs, 0)
		queryJobs(t, j, s.ExecutingJobs, 1)
		queryJobs(t, j, s.LatestFinalJobs, 0)
		queryJobs(t, j, s.LatestSuccessfulJobs, 0)

		err = s.Finalize(j.ID, ffs.Failed, errors.New("oops"), nil)
		require.NoError(t, err)

		queryJobs(t, j, s.QueuedJobs, 0)
		queryJobs(t, j, s.ExecutingJobs, 0)
		queryJobs(t, j, s.LatestFinalJobs, 1)
		queryJobs(t, j, s.LatestSuccessfulJobs, 0)
	})
	t.Run("ExecutingAndSucceeded", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob(t)

		err := s.Enqueue(j)
		require.NoError(t, err)

		dequeued, err := s.Dequeue()
		require.NoError(t, err)
		require.Equal(t, dequeued.ID, j.ID)

		err = s.Finalize(j.ID, ffs.Success, nil, nil)
		require.NoError(t, err)

		queryJobs(t, j, s.QueuedJobs, 0)
		queryJobs(t, j, s.ExecutingJobs, 0)
		queryJobs(t, j, s.LatestFinalJobs, 1)
		queryJobs(t, j, s.LatestSuccessfulJobs, 1)
	})
}

func queryJobs(t *testing.T, j ffs.StorageJob, f func(iid ffs.APIID, cids ...cid.Cid) []ffs.StorageJob, count int) {
	jobs := f(ffs.EmptyInstanceID)
	require.Len(t, jobs, count)
	jobs = f(j.APIID)
	require.Len(t, jobs, count)
	jobs = f(ffs.EmptyInstanceID, j.Cid)
	require.Len(t, jobs, count)
	jobs = f(j.APIID, j.Cid)
	require.Len(t, jobs, count)
	jobs = f(ffs.APIID("nope"))
	require.Empty(t, jobs)
	c, err := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs9y")
	require.NoError(t, err)
	jobs = f(ffs.EmptyInstanceID, c)
	require.Empty(t, jobs)
}

func createJob(t *testing.T) ffs.StorageJob {
	c, err := util.CidFromString("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	require.NoError(t, err)

	return ffs.StorageJob{
		ID:    ffs.NewJobID(),
		APIID: "ApiIDTest",
		Cid:   c,
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds)
	require.NoError(t, err)
	return store
}
