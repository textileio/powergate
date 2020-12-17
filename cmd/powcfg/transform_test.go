package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger2"
	logger "github.com/ipfs/go-log/v2"
	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/ffs"
)

func TestIpfsAddBump(t *testing.T) {
	logger.SetDebugLogging()
	_ = logger.SetLogLevel("badger", "error")

	for _, dryRun := range []bool{false, true} {
		dryRun := dryRun
		t.Run(fmt.Sprintf("dryrun %v", dryRun), func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			err := cp.Copy("testdata/badgerdumpv1", tmpDir+"/badgerdump")
			require.NoError(t, err)

			opts := &badger.DefaultOptions
			ds, err := badger.NewDatastore(tmpDir+"/badgerdump", opts)
			require.NoError(t, err)
			defer func() { require.NoError(t, ds.Close()) }()

			modified, err := applyTransform(ds, dryRun, bumpIpfsAddTimeout(12345))
			require.NoError(t, err)

			// Verify
			require.Equal(t, 678, modified)

			var cfg ffs.StorageConfig
			k := datastore.NewKey("/ffs/manager/api/274cb2c5-a0da-49fa-91a8-97b9d387d4fa/istore/cidstorageconfig/QmbtfAvVRVgEa9RH6vYpKg1HWbBQjxiZ6vNG1AAvt3vyfR")
			buf, err := ds.Get(k)
			require.NoError(t, err)
			err = json.Unmarshal(buf, &cfg)
			require.NoError(t, err)

			if dryRun {
				require.NotEqual(t, 12345, cfg.Hot.Ipfs.AddTimeout)
				return
			}
			require.Equal(t, 12345, cfg.Hot.Ipfs.AddTimeout)
		})
	}
}
