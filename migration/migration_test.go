package migration

import (
	"fmt"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

func TestEmptyDatastore(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return fmt.Errorf("This migration shouldn't be run on empty database")
		},
	}
	m := newMigrator(ms, true)

	v, empty, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.True(t, empty)

	err = m.Ensure()
	require.NoError(t, err)

	v, _, err = m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 1, v)
}

func TestNonEmptyDatastore(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms, false)

	v, empty, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, empty)

	err = m.Ensure()
	require.NoError(t, err)

	v, _, err = m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 1, v)
}

func TestNoop(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms, false)

	err := m.Ensure()
	require.NoError(t, err)

	err = m.Ensure()
	require.NoError(t, err)
}

func TestFailingMigration(t *testing.T) {
	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return fmt.Errorf("I failed")
		},
	}
	m := newMigrator(ms, false)

	v, _, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)

	err = m.Ensure()
	require.Error(t, err)

	v, _, err = m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)
}

func newMigrator(migrations map[int]Migration, empty bool) *Migrator {
	ds := tests.NewTxMapDatastore()
	m := New(ds, migrations)

	if !empty {
		_ = ds.Put(datastore.NewKey("foo"), []byte("bar"))
	}

	return m
}
