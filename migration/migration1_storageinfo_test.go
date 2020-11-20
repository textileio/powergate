package migration

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestV1_StorageInfo(t *testing.T) {
	ds := tests.NewTxMapDatastore()

	pre(t, ds, "testdata/v1_StorageInfo.pre")

	c1, _ := util.CidFromString("QmbtfAvVRVgEa9RH6vYpKg1HWbBQjxiZ6vNG1AAvt3vyf1")
	c2, _ := util.CidFromString("QmZTMaDfCMWqhUXYDnKup8ctCTyPxnriYW7G4JR8KXoX5M")
	cidOwners := map[cid.Cid][]ffs.APIID{
		c1: {ffs.APIID("ID1"), ffs.APIID("ID2")},
		c2: {ffs.APIID("ID2"), ffs.APIID("ID3")},
	}
	txn, _ := ds.NewTransaction(false)
	err := migrateStorageInfo(txn, cidOwners)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	post(t, ds, "testdata/v1_StorageInfo.post")
}

func pre(t *testing.T, ds datastore.TxnDatastore, path string) {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), ",", 2)
		err = ds.Put(datastore.NewKey(parts[0]), []byte(parts[1]))
		require.NoError(t, err)
	}
}

func post(t *testing.T, ds datastore.TxnDatastore, path string) {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	current := map[string][]byte{}
	q := query.Query{}
	res, err := ds.Query(q)
	require.NoError(t, err)
	defer func() { _ = res.Close() }()
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
	// Include the extra "version" key in datastore.
	expected["version"] = []byte("{\"Version\":1}")

	require.Equal(t, len(expected), len(current))
	for k1, v1 := range current {
		v2, ok := expected[k1]
		require.True(t, ok)
		require.Equal(t, v2, v1)
	}
}
