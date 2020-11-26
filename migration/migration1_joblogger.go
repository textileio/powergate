package migration

import (
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

func migrateJobLogger(ds datastoreReaderWriter, cidOwners map[cid.Cid][]ffs.APIID) error {
	q := query.Query{Prefix: "/ffs/joblogger"}
	res, err := ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying joblogger: %s", err)
	}
	defer func() { _ = res.Close() }()

	var lock sync.Mutex
	var errors []string
	lim := make(chan struct{}, 1000)
	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}
		lim <- struct{}{}

		r := r
		go func() {
			defer func() { <-lim }()

			originalKey := datastore.NewKey(r.Key)
			cidStr := originalKey.Namespaces()[2] // /ffs/joblogger/<cid>/<timestamp>
			cid, err := util.CidFromString(cidStr)
			if err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("parsing cid %s: %s", cidStr, err))
				lock.Unlock()
				return
			}
			timestampStr := originalKey.Namespaces()[3]

			owners := cidOwners[cid]

			// Step 1/2:
			// For each cid owner, we create the same registry
			// in the new key version.
			for _, iid := range owners {
				newKey := datastore.NewKey("/ffs/joblogger_v2/").ChildString(iid.String()).ChildString(cidStr).ChildString(timestampStr)
				if err := ds.Put(newKey, r.Value); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("copying job log: %s", err))
					lock.Unlock()
					return
				}
			}

			// Step 2/2:
			// Delete old datastore key.
			if err := ds.Delete(originalKey); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("deleting old key: %s", err))
				lock.Unlock()
				return
			}
		}()
	}

	for i := 0; i < len(lim); i++ {
		lim <- struct{}{}
	}

	if len(errors) > 0 {
		for _, m := range errors {
			log.Error(m)
		}
		return fmt.Errorf("migration had %d errors", len(errors))
	}

	return nil
}
