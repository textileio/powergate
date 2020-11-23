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

func migrateStorageInfo(txn datastore.Txn, cidOwners map[cid.Cid][]ffs.APIID) error {
	q := query.Query{Prefix: "/ffs/scheduler/cistore"}
	res, err := txn.Query(q)
	if err != nil {
		return fmt.Errorf("querying cistore: %s", err)
	}
	defer func() { _ = res.Close() }()

	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}

		originalKey := datastore.NewKey(r.Key)
		cidStr := originalKey.Namespaces()[3] // /ffs/scheduler/cistore/<cid>
		cid, err := util.CidFromString(cidStr)
		if err != nil {
			return fmt.Errorf("discovered invalid cid: %s", err)
		}

		owners := cidOwners[cid]
		// Step 1/2:
		// For each cid owner, we create the same registry
		// prexifing the iid considering the new namespace structure.
		for _, iid := range owners {
			log.Infof("Migrating storageinfo to owner %s...", iid)

			var si ffs.StorageInfo
			if err := json.Unmarshal(r.Value, &si); err != nil {
				return fmt.Errorf("unmarshaling storageconfig: %s", err)
			}
			si.APIID = iid
			buf, err := json.Marshal(si)
			if err != nil {
				return fmt.Errorf("marshaling now entry: %s", err)
			}

			newKey := datastore.NewKey("/ffs/scheduler/cistore").ChildString(iid.String()).ChildString(cidStr)
			if err := txn.Put(newKey, buf); err != nil {
				return fmt.Errorf("copying storageinfo: %s", err)
			}
		}

		// Step 2/2:
		// Delete old datastore key.
		log.Infof("Deleting original storageinfo key...")
		if err := txn.Delete(originalKey); err != nil {
			return fmt.Errorf("deleting old key: %s", err)
		}
	}

	return nil
}
