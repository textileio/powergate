package migration

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
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

func TestForwardOnly(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: func(_ datastore.Txn) error {
			return nil
		},
	}
	m := newMigrator(ms, false)
	m.bootstrapEmptyDatastore(10)

	v, _, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 10, v)

	err = m.Ensure()
	require.Error(t, err)
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

func pre(t *testing.T, ds datastore.TxnDatastore, path string) {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()

	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), ",", 2)
		err = ds.Put(datastore.NewKey(parts[0]), []byte(parts[1]))
		require.NoError(t, err)
	}
}

func post(t *testing.T, ds datastore.TxnDatastore, path string) {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()

	current := map[string][]byte{}
	q := query.Query{}
	res, err := ds.Query(q)
	require.NoError(t, err)
	defer func() { require.NoError(t, res.Close()) }()
	for r := range res.Next() {
		require.NoError(t, r.Error)
		current[r.Key] = r.Value
	}

	expected := map[string][]byte{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), ",", 2)
		expected[parts[0]] = []byte(parts[1])
	}

	require.Equal(t, len(expected), len(current))
	for k1, v1 := range current {
		v2, ok := expected[k1]
		require.True(t, ok)
		if !bytes.Equal(v2, v1) {
			fmt.Printf("HUH1: %s\n", string(v2))
			fmt.Printf("HUH2: %s\n", string(v1))
		}
		require.Equal(t, v2, v1)
	}
}
