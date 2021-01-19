package migration

import (
	"encoding/json"
	"fmt"
	"sync"

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
	UseTxn: false,
	Run: func(ds datastoreReaderWriter) error {
		q := query.Query{Prefix: "/ffs/scheduler/sjstore/job"}
		res, err := ds.Query(q)
		if err != nil {
			return fmt.Errorf("querying sjstore jobs: %s", err)
		}
		defer func() { _ = res.Close() }()

		var count int
		var lock sync.Mutex
		var errors []string
		lim := make(chan struct{}, 1000)
		for r := range res.Next() {
			if r.Error != nil {
				return fmt.Errorf("iterating results: %s", r.Error)
			}

			lim <- struct{}{}

			r := r
			go func() {
				defer func() { <-lim }()

				var job storageJob
				if err := json.Unmarshal(r.Value, &job); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("unmarshaling job: %s", err))
					lock.Unlock()
					return
				}

				apiidKey := datastore.NewKey("/ffs/scheduler/sjstore/apiid").ChildString(job.APIID).ChildString(job.Cid.String()).ChildString(fmt.Sprintf("%d", job.CreatedAt))
				cidKey := datastore.NewKey("/ffs/scheduler/sjstore/cid").ChildString(job.Cid.String()).ChildString(job.APIID).ChildString(fmt.Sprintf("%d", job.CreatedAt))

				if err := ds.Put(apiidKey, []byte(job.ID)); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("putting apiid index record in datastore: %s", err))
					lock.Unlock()
					return

				}
				if err := ds.Put(cidKey, []byte(job.ID)); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("putting cid index record in datastore: %s", err))
					lock.Unlock()
					return

				}

			}()
			count++
		}

		for i := 0; i < cap(lim); i++ {
			lim <- struct{}{}
		}

		if len(errors) > 0 {
			for _, m := range errors {
				log.Error(m)
			}
			return fmt.Errorf("migration had %d errors", len(errors))
		}

		log.Infof("migration indexed %d storage-jobs", count)

		return nil
	},
}
