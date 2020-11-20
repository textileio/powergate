package migration

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

func migrateJobLogger(txn datastore.Txn, cidOwners map[cid.Cid][]ffs.APIID) error {
	q := query.Query{Prefix: "/ffs/joblogger"}
	res, err := txn.Query(q)
	if err != nil {
		return fmt.Errorf("querying joblogger: %s", err)
	}
	defer func() { _ = res.Close() }()

	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}

		originalKey := datastore.NewKey(r.Key)
		cidStr := originalKey.Namespaces()[2] // /ffs/joblogger/<cid>/<timestamp>
		cid, err := util.CidFromString(cidStr)
		if err != nil {
			return fmt.Errorf("discovered invalid cid: %s", err)
		}
		timestampStr := originalKey.Namespaces()[3]

		owners := cidOwners[cid]

		// Step 1/2:
		// For each cid owner, we create the same registry
		// adding the new field APIID to the corresponding
		// owner.
		for _, iid := range owners {
			log.Infof("Migrating job log to owner %s...", iid)

			newKey := datastore.NewKey("/ffs/jobloger").ChildString(iid).ChildString(cidStr).ChildString(timestampStr)
			if err := txn.Put(newKey, r.Value); err != nil {
				return fmt.Errorf("copying job log: %s", err)
			}
		}

		// Step 2/2:
		// Delete old datastore key.
		log.Infof("Deleting original log key...")
		if err := txn.Delete(originalKey); err != nil {
			return fmt.Errorf("deleting old key: %s", err)
		}
	}

	return nil
}
