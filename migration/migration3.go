package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

type storageJob struct {
	ID        string
	APIID     string
	Cid       cid.Cid
	CreatedAt int64
}

// V3StorageJobsIndexMigration contains the logic to upgrade a datastore from
// version 2 to version 3.
var V3StorageJobsIndexMigration = Migration{
	UseTxn: true,
	Run: func(ds datastoreReaderWriter) error {
		q := query.Query{Prefix: "/ffs/scheduler/sjstore/job"}
		res, err := ds.Query(q)
		if err != nil {
			return fmt.Errorf("querying sjstore jobs: %s", err)
		}
		defer func() { _ = res.Close() }()

		for r := range res.Next() {
			if r.Error != nil {
				return fmt.Errorf("iterating results: %s", r.Error)
			}

			var job storageJob
			if err := json.Unmarshal(r.Value, &job); err != nil {
				return fmt.Errorf("unmarshaling job: %s", err)
			}

			apiidKey := datastore.NewKey("/ffs/scheduler/sjstore/apiid").ChildString(job.APIID).ChildString(job.Cid.String()).ChildString(fmt.Sprintf("%d", job.CreatedAt))
			cidKey := datastore.NewKey("/ffs/scheduler/sjstore/cid").ChildString(job.Cid.String()).ChildString(job.APIID).ChildString(fmt.Sprintf("%d", job.CreatedAt))

			if err := ds.Put(apiidKey, []byte(job.ID)); err != nil {
				return fmt.Errorf("putting apiid index record in datastore: %s", err)
			}
			if err := ds.Put(cidKey, []byte(job.ID)); err != nil {
				return fmt.Errorf("putting cid index record in datastore: %s", err)
			}
		}
		return nil
	},
}
