package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/util"
)

type v0trackedStorageConfig struct {
	IID           ffs.APIID
	StorageConfig ffs.StorageConfig
}

// In v1 we persist the same struct but as an array, instead
// of a single element.
type v1trackedStorageConfig v0trackedStorageConfig

func migrateTrackstore(ds datastoreReaderWriter, cidOwners map[cid.Cid][]ffs.APIID) error {
	q := query.Query{Prefix: "/ffs/scheduler/tstore"}
	res, err := ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying tstore: %s", err)
	}
	defer func() { _ = res.Close() }()

	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}

		originalKey := datastore.NewKey(r.Key)
		cidStr := originalKey.Namespaces()[3] // /ffs/scheduler/tstore/<cid>
		cid, err := util.CidFromString(cidStr)
		if err != nil {
			return fmt.Errorf("discovered invalid cid: %s", err)
		}

		// Step 1/2:
		// For each cid owner, we replace the registry
		// with a slice of owners.
		owners := cidOwners[cid]
		newEntry := make([]v1trackedStorageConfig, len(owners))
		for i, iid := range owners {
			sg, err := v0GetStorageConfig(ds, iid, cid)
			if err != nil {
				return fmt.Errorf("get storage config of cid from owner: %s", err)
			}

			if sg.Repairable || (sg.Cold.Enabled && sg.Cold.Filecoin.Renew.Enabled) {
				newEntry[i] = v1trackedStorageConfig{IID: iid, StorageConfig: sg}
			}
		}

		// Step 2/2:
		// Replace original entry with new entry.
		buf, err := json.Marshal(newEntry)
		if err != nil {
			return fmt.Errorf("marshaling new entry: %s", err)
		}

		if err := ds.Put(originalKey, buf); err != nil {
			return fmt.Errorf("put new entry in datastore: %s", err)
		}
	}

	return nil
}
