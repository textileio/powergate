package jstore

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
)

func TestEnqueue(t *testing.T) {
	t.Parallel()
	s := create(t)
	j := createJob()
	err := s.Enqueue(j)
	require.Nil(t, err)
	jQueued, err := s.Get(j.ID)
	require.Nil(t, err)
	j.Status = ffs.Queued
	require.Equal(t, j, jQueued)
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob()
		err := s.Enqueue(j)
		require.Nil(t, err)
		j2, err := s.Dequeue()
		require.Nil(t, err)
		j.Status = ffs.Executing
		require.NotNil(t, j2)
		require.Equal(t, j, *j2)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j, err := s.Dequeue()
		require.Nil(t, err)
		require.Nil(t, j)
	})
	t.Run("ExecutingAndFinalized", func(t *testing.T) {
		t.Parallel()
		s := create(t)

		j1 := createJob()
		err := s.Enqueue(j1)
		require.Nil(t, err)
		_, err = s.Dequeue()
		require.Nil(t, err)

		j2 := createJob()
		err = s.Enqueue(j2)
		require.Nil(t, err)

		// Dequeue should be nil since j1 is Executing,
		// and j2 can't be returned until that fininishes since
		// both are jobs for the same Cid.
		jq, err := s.Dequeue()
		require.Nil(t, err)
		require.Nil(t, jq)

		errors := []ffs.DealError{ffs.DealError{ProposalCid: cid.Undef, Miner: "t0100", Message: "abcd"}}
		err = s.Finalize(j1.ID, ffs.Success, nil, errors)
		require.Nil(t, err)
		// Dequeue now returns a new job since the Executing one
		// was finalized
		jq, err = s.Dequeue()
		require.Nil(t, err)
		require.NotNil(t, jq)
		j2.Status = ffs.Executing
		require.Equal(t, j2, *jq)
	})
}

func TestCancelation(t *testing.T) {
	t.Parallel()
	s := create(t)

	j1 := createJob()
	err := s.Enqueue(j1)
	require.Nil(t, err)

	j2 := createJob()
	err = s.Enqueue(j2)
	require.Nil(t, err)

	j1Canceled, err := s.Get(j1.ID)
	require.Nil(t, err)
	require.Equal(t, ffs.Canceled, j1Canceled.Status)

	j2Queued, err := s.Get(j2.ID)
	require.Nil(t, err)
	require.Equal(t, ffs.Queued, j2Queued.Status)
}

func createJob() ffs.Job {
	c, err := cid.Decode("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	if err != nil {
		panic(err)
	}
	return ffs.Job{
		ID:    ffs.NewJobID(),
		APIID: "ApiIDTest",
		Cid:   c,
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	return New(ds)

}
