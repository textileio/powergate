package migration

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger2"
	"github.com/stretchr/testify/require"
	mongods "github.com/textileio/go-ds-mongo"
	"github.com/textileio/powergate/v2/tests"

	logger "github.com/ipfs/go-log/v2"
)

func TestEmptyDatastore(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: {
			UseTxn: true,
			Run: func(_ datastoreReaderWriter) error {
				return fmt.Errorf("This migration shouldn't be run on empty database")
			},
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
		1: {
			UseTxn: true,
			Run: func(_ datastoreReaderWriter) error {
				return nil
			},
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
		1: {
			UseTxn: true,
			Run: func(_ datastoreReaderWriter) error {
				return nil
			},
		},
	}

	m := newMigrator(ms, false)
	err := m.bootstrapEmptyDatastore(10)
	require.NoError(t, err)

	v, _, err := m.getCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, 10, v)

	err = m.Ensure()
	require.Error(t, err)
}

func TestNoop(t *testing.T) {
	t.Parallel()

	ms := map[int]Migration{
		1: {
			UseTxn: true,
			Run: func(_ datastoreReaderWriter) error {
				return nil
			},
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
		1: {
			UseTxn: true,
			Run: func(_ datastoreReaderWriter) error {
				return fmt.Errorf("I failed")
			},
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

func TestRealDataBadgerFromV0(t *testing.T) {
	logger.SetDebugLogging()
	_ = logger.SetLogLevel("badger", "error")
	tmpDir := t.TempDir()
	err := copyDir("testdata/badgerdumpv0", tmpDir+"/badgerdump")
	require.NoError(t, err)

	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(tmpDir+"/badgerdump", opts)
	require.NoError(t, err)
	defer func() { require.NoError(t, ds.Close()) }()

	migrations := map[int]Migration{
		1: V1MultitenancyMigration,
	}
	m := New(ds, migrations)
	err = m.Ensure()
	require.NoError(t, err)
}

func TestRealDataBadgerFromV1(t *testing.T) {
	logger.SetDebugLogging()
	_ = logger.SetLogLevel("badger", "error")
	tmpDir := t.TempDir()
	err := copyDir("testdata/badgerdumpv1", tmpDir+"/badgerdump")
	require.NoError(t, err)

	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(tmpDir+"/badgerdump", opts)
	require.NoError(t, err)
	defer func() { require.NoError(t, ds.Close()) }()

	migrations := map[int]Migration{
		1: V1MultitenancyMigration,
		2: V2StorageInfoDealIDs,
	}
	m := New(ds, migrations)
	err = m.Ensure()
	require.NoError(t, err)
}

func TestRealDataBadgerFromV2(t *testing.T) {
	logger.SetDebugLogging()
	_ = logger.SetLogLevel("badger", "error")
	tmpDir := t.TempDir()
	err := copyDir("testdata/badgerdumpv1", tmpDir+"/badgerdump")
	require.NoError(t, err)

	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(tmpDir+"/badgerdump", opts)
	require.NoError(t, err)
	defer func() { require.NoError(t, ds.Close()) }()

	migrations := map[int]Migration{
		1: V1MultitenancyMigration,
		2: V2StorageInfoDealIDs,
		3: V3StorageJobsIndexMigration,
	}
	m := New(ds, migrations)
	err = m.Ensure()
	require.NoError(t, err)
}

func TestRealDataBadgerFromV3(t *testing.T) {
	logger.SetDebugLogging()
	_ = logger.SetLogLevel("badger", "error")
	tmpDir := t.TempDir()
	err := copyDir("testdata/badgerdumpv1", tmpDir+"/badgerdump")
	require.NoError(t, err)

	opts := &badger.DefaultOptions
	ds, err := badger.NewDatastore(tmpDir+"/badgerdump", opts)
	require.NoError(t, err)
	defer func() { require.NoError(t, ds.Close()) }()

	migrations := map[int]Migration{
		1: V1MultitenancyMigration,
		2: V2StorageInfoDealIDs,
		3: V3StorageJobsIndexMigration,
		4: V4RecordsMigration,
	}
	m := New(ds, migrations)
	err = m.Ensure()
	require.NoError(t, err)
}

func TestRealDataRemoteMongo(t *testing.T) {
	t.SkipNow()
	logger.SetDebugLogging()
	mongoCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	mongoURI := ""
	mongoDB := ""
	ds, err := mongods.New(mongoCtx, mongoURI, mongoDB, mongods.WithCollName("migration_kvstore"), mongods.WithOpTimeout(time.Hour), mongods.WithTxnTimeout(time.Hour))
	require.NoError(t, err)
	defer func() { require.NoError(t, ds.Close()) }()

	migrations := map[int]Migration{
		1: V1MultitenancyMigration,
		2: V2StorageInfoDealIDs,
	}
	m := New(ds, migrations)
	err = m.Ensure()
	require.NoError(t, err)
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
		require.Equal(t, v2, v1)
	}
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
