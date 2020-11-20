package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

type v0_trackedStorageConfig struct {
	IID           ffs.APIID
	StorageConfig ffs.StorageConfig
}

// In v1 we persist the same struct but as an array, instead
// of a single element.
type v1_trackedStorageConfig v0_trackedStorageConfig

func migrateTrackstore(txn datastore.Txn, cidOwners map[cid.Cid][]ffs.APIID) error {
	q := query.Query{Prefix: "/ffs/scheduler/tstore"}
	res, err := txn.Query(q)
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
		log.Infof("Migrating trackstore entry of cid %s...", cid)

		// Step 1/2:
		// For each cid owner, we replace the registry
		// with a slice of owners.
		owners := cidOwners[cid]
		newEntry := make([]v1_trackedStorageConfig, len(owners))
		for _, iid := range owners {
			sg, err := v0_GetStorageConfig(txn, iid, cid)
			if err != nil {
				return fmt.Errorf("get storage config of cid from owner: %s", err)
			}

			if sg.Repairable || (sg.Cold.Enabled && sg.Cold.Filecoin.Renew.Enabled) {
				newOwnerEntry := v1_trackedStorageConfig{IID: iid, StorageConfig: sg}
				newEntry = append(newEntry, newOwnerEntry)
			}
		}

		// Step 2/2:
		// Replace original entry with new entry.
		buf, err := json.Marshal(newEntry)
		if err != nil {
			return fmt.Errorf("marshaling new entry: %s", err)
		}
		if err := txn.Put(originalKey, buf); err != nil {
			return fmt.Errorf("put new entry in datastore: %s", err)
		}

		log.Infof("Migrated trackstore entry with %d owners...", len(newEntry))
	}

	return nil
}
