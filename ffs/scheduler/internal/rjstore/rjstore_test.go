package rjstore

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
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
}

func createJob(t *testing.T) ffs.Job {
	return ffs.Job{
		ID:          ffs.NewJobID(),
		APIID:       "ApiIDTest",
		RetrievalID: ffs.NewRetrievalID(),
	}
}

func create(t *testing.T) *Store {
	ds := tests.NewTxMapDatastore()
	store, err := New(ds)
	require.NoError(t, err)
	return store
}
