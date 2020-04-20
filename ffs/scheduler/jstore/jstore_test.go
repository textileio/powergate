package jstore

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
)

func TestQueue(t *testing.T) {
	t.Parallel()
	s := create(t)
	j := createJob()
	err := s.Queue(j)
	require.Nil(t, err)
	j_queued, err := s.Get(j.ID)
	require.Nil(t, err)
	j.Status = ffs.Queued
	require.Equal(t, j, j_queued)
}

func TestDequeue(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j := createJob()
		err := s.Queue(j)
		require.Nil(t, err)
		j2, err := s.Dequeue()
		require.Nil(t, err)
		j.Status = ffs.InProgress
		require.Equal(t, j, *j2)
	})
	t.Run("Empty", func(t *testing.T) {
		t.Parallel()
		s := create(t)
		j, err := s.Dequeue()
		require.Nil(t, err)
		require.Nil(t, j)
	})
}

func TestCancelation(t *testing.T) {
	t.Parallel()
	s := create(t)

	j1 := createJob()
	err := s.Queue(j1)
	require.Nil(t, err)

	j2 := createJob()
	err = s.Queue(j2)
	require.Nil(t, err)

	j1_canceled, err := s.Get(j1.ID)
	require.Nil(t, err)
	require.Equal(t, ffs.Canceled, j1_canceled.Status)

	j2_queued, err := s.Get(j2.ID)
	require.Nil(t, err)
	require.Equal(t, ffs.Queued, j2_queued.Status)
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
