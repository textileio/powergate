package migration

import (
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/v2/util"
)

func migrateStartedDeals(ds datastoreReaderWriter) error {
	executingJobs, err := v0GetExecutingJobs(ds)
	if err != nil {
		return fmt.Errorf("getting executing jobs: %s", err)
	}
	q := query.Query{Prefix: "/ffs/scheduler/sjstore/starteddeals"}
	res, err := ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying started deals in sjstore: %s", err)
	}
	defer func() { _ = res.Close() }()

	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}

		originalKey := datastore.NewKey(r.Key)
		cidStr := originalKey.Namespaces()[4] // /ffs/scheduler/sjstore/starteddeals/<cid>
		cid, err := util.CidFromString(cidStr)
		if err != nil {
			return fmt.Errorf("discovered invalid cid: %s", err)
		}

		owner, ok := executingJobs[cid]
		if ok {
			// Step 1/2:
			// Add corresponding iid in started deals key namespace.
			newKey := datastore.NewKey("/ffs/scheduler/sjstore/starteddeals_v2").ChildString(owner.String()).ChildString(cidStr)
			if err := ds.Put(newKey, r.Value); err != nil {
				return fmt.Errorf("copying started deal to new key: %s", err)
			}
		}

		// Step 2/2:
		// Delete old datastore key.
		if err := ds.Delete(originalKey); err != nil {
			return fmt.Errorf("deleting old key: %s", err)
		}
	}

	return nil
}
