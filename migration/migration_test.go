package migration

import (
	"fmt"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

func TestBoostrapMigration(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms)

	v, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)

	err = m.Ensure(1)
	require.NoError(t, err)

	v, err = m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 1, v)
}

func TestMissingMigration(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms)

	err := m.Ensure(2)
	require.Error(t, err)
}

func TestWrongTarget(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms)

	err := m.Ensure(-1)
	require.Error(t, err)
}

func TestNoop(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms)

	err := m.Ensure(1)
	require.NoError(t, err)

	err = m.Ensure(1)
	require.NoError(t, err)
}

func TestFailingMigration(t *testing.T) {
	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return fmt.Errorf("I failed")
		},
	}
	m := newMigrator(ms)

	v, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)

	err = m.Ensure(1)
	require.Error(t, err)

	v, err = m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 0, v)
}

func newMigrator(migrations map[int]Migration) *Migrator {
	ds := tests.NewTxMapDatastore()
	m := New(ds, migrations)

	return m
}
