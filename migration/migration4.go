package migration

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/util"
)

// V4Records contains the logic to upgrade a datastore from
// version 3 to version 4.
var V4RecordsMigration = Migration{
	UseTxn: false,
	Run: func(ds datastoreReaderWriter) error {
		if err := v4IncludeIDInRetrievalRecords(ds); err != nil {
			return fmt.Errorf("including ids in retrieval records: %s", err)
		}

		if err := v4PopulateUpdatedAtIndex(ds); err != nil {
			return fmt.Errorf("populating updated at index: %s", err)
		}

		return nil
	},
}

type v4PartialRecordTime struct {
	Time int64
}

func v4PopulateUpdatedAtIndex(ds datastoreReaderWriter) error {
	prefixes := []string{"/deals/storage-pending", "/deals/storage-final", "/deals/retrieval"}

	for _, prefix := range prefixes {
		q := query.Query{Prefix: prefix}
		res, err := ds.Query(q)
		if err != nil {
			return fmt.Errorf("querying records: %s", err)
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

				var rec v4PartialRecordTime
				if err := json.Unmarshal(r.Value, &rec); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("unmarshaling record: %s", err))
					lock.Unlock()
					return
				}

				// This shouldn't happen.
				if rec.Time == 0 {
					lock.Lock()
					errors = append(errors, "time is zero")
					lock.Unlock()
					return
				}

				// The index stores at the nano precision.
				updatedAtTime := time.Unix(rec.Time, 0).UnixNano()

				var indexKey datastore.Key
				if strings.HasPrefix(r.Key, "/deals/storage-") {
					indexKey = datastore.NewKey("/deals/updatedatidx/storage").ChildString(strconv.FormatInt(updatedAtTime, 10))
				} else if strings.HasPrefix(r.Key, "/deals/retrieval") {
					indexKey = datastore.NewKey("/deals/updatedatidx/retrieval").ChildString(strconv.FormatInt(updatedAtTime, 10))

				} else {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("unkown prefix: %s", r.Key))
					lock.Unlock()
					return
				}

				if err := ds.Put(indexKey, []byte(r.Key)); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("saving updated-at index: %s", err))
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

		log.Infof("migration indexed %d %s records", count, prefix)
	}
	return nil
}

func v4IncludeIDInRetrievalRecords(ds datastoreReaderWriter) error {
	q := query.Query{Prefix: "/deals/retrieval"}
	res, err := ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying retrieval records: %s", err)
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

			var rr deals.RetrievalDealRecord
			if err := json.Unmarshal(r.Value, &rr); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("unmarshaling retrieval record: %s", err))
				lock.Unlock()
				return
			}

			rr.ID = v4RetrievalID(rr)
			buf, err := json.Marshal(rr)
			if err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("marshaling retrieval record: %s", err))
				lock.Unlock()
				return
			}

			if err := ds.Put(datastore.NewKey(r.Key), buf); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("put retrieval record: %s", err))
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

	log.Infof("migration populated %d retrieval records ids", count)

	return nil
}

func v4RetrievalID(rr deals.RetrievalDealRecord) string {
	str := fmt.Sprintf("%v%v%v%v", rr.Time, rr.Addr, rr.DealInfo.Miner, util.CidToString(rr.DealInfo.RootCid))
	sum := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", sum[:])
}
