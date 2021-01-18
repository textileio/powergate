package migration

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
)

// V2StorageInfoDealIDs contains the logic to upgrade a datastore from
// version 1 to version 2. Transactionality is disabled since is a big migration.
var V2StorageInfoDealIDs = Migration{
	UseTxn: false,
	Run: func(ds datastoreReaderWriter) error {
		propDealIDMap, err := v2GenProposalCidToDealID(ds)
		if err != nil {
			return fmt.Errorf("getting cid owners: %s", err)
		}
		log.Infof("Starting DealID filling...")
		if err := v2MigrationDealIDFilling(ds, propDealIDMap); err != nil {
			return fmt.Errorf("migrating job logger: %s", err)
		}
		log.Infof("DealID filling migration finished")

		return nil
	},
}

func v2GenProposalCidToDealID(ds datastoreReaderWriter) (map[cid.Cid]uint64, error) {
	ret := map[cid.Cid]uint64{}
	qs := []query.Query{{Prefix: "/deals/storage-pending"}, {Prefix: "/deals/storage-final"}}

	for _, q := range qs {
		var lock sync.Mutex
		var errors []string
		lim := make(chan struct{}, 1000)

		res, err := ds.Query(q)
		if err != nil {
			return nil, fmt.Errorf("querying storage info: %s", err)
		}
		defer func() {
			_ = res.Close()
		}()

		for r := range res.Next() {
			if r.Error != nil {
				return nil, fmt.Errorf("query result: %s", r.Error)
			}
			lim <- struct{}{}

			r := r
			go func() {
				defer func() { <-lim }()
				var dr deals.StorageDealRecord
				if err := json.Unmarshal(r.Value, &dr); err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("unmarshaling query result: %s", err))
					lock.Unlock()
					return
				}
				if dr.DealInfo.DealID != 0 {
					lock.Lock()
					ret[dr.DealInfo.ProposalCid] = dr.DealInfo.DealID
					lock.Unlock()
				}
			}()
		}
		for i := 0; i < cap(lim); i++ {
			lim <- struct{}{}
		}

		if len(errors) > 0 {
			for _, m := range errors {
				log.Error(m)
			}
			return nil, fmt.Errorf("map build had %d errors", len(errors))
		}
	}

	return ret, nil
}

type v1PartialStorageInfo struct {
	Cold v1PartialCold
}

type v1PartialCold struct {
	Filecoin v1PartialFilecoinInfo
}

type v1PartialFilecoinInfo struct {
	Proposals []v1PartialFilStorage
}

type v1PartialFilStorage struct {
	ProposalCid cid.Cid
}

func v2MigrationDealIDFilling(ds datastoreReaderWriter, propsDealID map[cid.Cid]uint64) error {
	log.Infof("Total map size: %d", len(propsDealID))
	q := query.Query{Prefix: "/ffs/scheduler/cistore_v2"} // /ffs/scheduler/cistore_v2/<iid>/<cid>
	res, err := ds.Query(q)
	if err != nil {
		return fmt.Errorf("querying storage info store: %s", err)
	}
	defer func() { _ = res.Close() }()

	var lock sync.Mutex
	var errors []string
	lim := make(chan struct{}, 1000)

	var cantBeMapped, canBeMapped int
	for r := range res.Next() {
		if r.Error != nil {
			return fmt.Errorf("iterating result: %s", r.Error)
		}
		lim <- struct{}{}

		r := r
		go func() {
			defer func() { <-lim }()

			var oldCi v1PartialStorageInfo
			var newCi ffs.StorageInfo
			if err := json.Unmarshal(r.Value, &oldCi); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("unmarshaling storage info from datastore: %s", err))
				lock.Unlock()
				return
			}
			if err := json.Unmarshal(r.Value, &newCi); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("unmarshaling storage info from datastore: %s", err))
				lock.Unlock()
				return
			}

			var filledProposals []ffs.FilStorage
			for i, p := range oldCi.Cold.Filecoin.Proposals {
				dealID, ok := propsDealID[p.ProposalCid]
				if !ok {
					lock.Lock()
					cantBeMapped++
					lock.Unlock()
					continue
				}
				lock.Lock()
				canBeMapped++
				lock.Unlock()
				newCi.Cold.Filecoin.Proposals[i].DealID = dealID
				filledProposals = append(filledProposals, newCi.Cold.Filecoin.Proposals[i])
			}
			newCi.Cold.Filecoin.Proposals = filledProposals

			buf, err := json.Marshal(newCi)
			if err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("marshaling storage info for datastore: %s", err))
				lock.Unlock()
			}
			if err := ds.Put(datastore.NewKey(r.Key), buf); err != nil {
				lock.Lock()
				errors = append(errors, fmt.Sprintf("put storage info in datastore: %s", err))
				lock.Unlock()
			}
		}()
	}

	for i := 0; i < cap(lim); i++ {
		lim <- struct{}{}
	}

	log.Infof("Couldn't be mapped %d, Could be mapped %d", cantBeMapped, canBeMapped)

	if len(errors) > 0 {
		for _, m := range errors {
			log.Error(m)
		}
		return fmt.Errorf("migration had %d errors", len(errors))
	}

	return nil
}
