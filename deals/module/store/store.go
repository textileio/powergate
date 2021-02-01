package store

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/util"
)

var (
	dsBaseStoragePending = datastore.NewKey("storage-pending")
	dsBaseStorageFinal   = datastore.NewKey("storage-final")
	dsBaseRetrieval      = datastore.NewKey("retrieval")

	// TODO: we need a purging daemon to delete index entries older than X days.
	dsStorageUpdatedAtIdx   = datastore.NewKey("updatedatidx/storage")
	dsRetrievalUpdatedAtIdx = datastore.NewKey("updatedatidx/retrieval")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("cid info not found")

	log = logging.Logger("deals-records")
)

// Store stores deal and retrieval records.
type Store struct {
	ds   datastore.TxnDatastore
	lock sync.Mutex
}

// New returns a new *Store.
func New(ds datastore.TxnDatastore) *Store {
	return &Store{
		ds: ds,
	}
}

// PutStorageDeal saves a storage deal record.
func (s *Store) PutStorageDeal(dr deals.StorageDealRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("creating transaction: %s", err)
	}
	defer txn.Discard()

	if dr.Pending && dr.ErrMsg != "" {
		return fmt.Errorf("pending storage records can't have error messages")
	}

	// If not pending, delete any saved record in the 'pending' keyspace.
	// If not exists, `Delete()` is a noop.
	if !dr.Pending {
		if err := txn.Delete(makePendingDealKey(dr.DealInfo.ProposalCid)); err != nil {
			return fmt.Errorf("delete pending deal storage record: %s", err)
		}
	}

	dr.UpdatedAt = time.Now().UnixNano()
	buf, err := json.Marshal(dr)
	if err != nil {
		return fmt.Errorf("marshaling storage deal record: %s", err)
	}

	var key datastore.Key
	if dr.Pending {
		key = makePendingDealKey(dr.DealInfo.ProposalCid)
	} else {
		key = makeFinalDealKey(dr.DealInfo.ProposalCid)
	}

	if err := txn.Put(key, buf); err != nil {
		return fmt.Errorf("put storage deal record: %s", err)
	}

	updatedAtIndexKey := makeStorageUpdatedAtIndexKey(dr.UpdatedAt)
	if err := txn.Put(updatedAtIndexKey, key.Bytes()); err != nil {
		return fmt.Errorf("saving updated-at index: %s", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %s", err)
	}

	return nil
}

// GetPendingStorageDeals returns all the pending storage deal records.
func (s *Store) GetPendingStorageDeals() ([]deals.StorageDealRecord, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := query.Query{Prefix: dsBaseStoragePending.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []deals.StorageDealRecord
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		var dr deals.StorageDealRecord
		if err := json.Unmarshal(r.Value, &dr); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, dr)
	}
	return ret, nil
}

// GetFinalStorageDeals returns all final storage deal records.
func (s *Store) GetFinalStorageDeals() ([]deals.StorageDealRecord, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := query.Query{Prefix: dsBaseStorageFinal.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []deals.StorageDealRecord
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		var dealRecord deals.StorageDealRecord
		if err := json.Unmarshal(r.Value, &dealRecord); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, dealRecord)
	}

	return ret, nil
}

// PutRetrieval saves a retrieval deal record.
func (s *Store) PutRetrieval(rr deals.RetrievalDealRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("creating transaction: %s", err)
	}
	defer txn.Discard()

	rr.UpdatedAt = time.Now().UnixNano()
	buf, err := json.Marshal(rr)
	if err != nil {
		return fmt.Errorf("marshaling RetrievalRecord: %s", err)
	}
	key := makeRetrievalKey(rr)
	if err := txn.Put(key, buf); err != nil {
		return fmt.Errorf("put RetrievalRecord: %s", err)
	}

	updatedAtIndexKey := makeRetrievalUpdatedAtIndexKey(rr.UpdatedAt)
	if err := txn.Put(updatedAtIndexKey, key.Bytes()); err != nil {
		return fmt.Errorf("saving updated-at index: %s", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %s", err)
	}

	return nil
}

// GetRetrievals returns all retrieval deal records.
func (s *Store) GetRetrievals() ([]deals.RetrievalDealRecord, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	q := query.Query{Prefix: dsBaseRetrieval.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []deals.RetrievalDealRecord
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		var rr deals.RetrievalDealRecord
		if err := json.Unmarshal(r.Value, &rr); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, rr)
	}
	return ret, nil
}

// GetUpdatedStorageDealRecordsSince returns all the storage deal records that got created or updated
// since sinceNano.
func (s *Store) GetUpdatedStorageDealRecordsSince(sinceNano int64, limit int) ([]deals.StorageDealRecord, error) {
	q := query.Query{
		Prefix: dsStorageUpdatedAtIdx.String(),
	}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing updated storage deal records query result: %s", err)
		}
	}()

	msdrs := make(map[datastore.Key]deals.StorageDealRecord)
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("getting query result: %s", r.Error)
		}
		unixNanoKey := datastore.NewKey(r.Key).Namespaces()[2]
		unixNano, err := strconv.ParseInt(unixNanoKey, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing unix nano key: %s", err)
		}

		if unixNano < sinceNano {
			continue
		}

		recordKey := datastore.RawKey(string(r.Value))
		buf, err := s.ds.Get(recordKey)
		if err != nil {
			return nil, fmt.Errorf("get updated storage deal record from store: %s", err)
		}
		var sr deals.StorageDealRecord
		if err := json.Unmarshal(buf, &sr); err != nil {
			return nil, fmt.Errorf("unmarshaling updated storage deal record from store: %s", err)
		}
		msdrs[recordKey] = sr

		limit--
		if limit == 0 {
			break
		}
	}

	ret := make([]deals.StorageDealRecord, 0, len(msdrs))
	for _, v := range msdrs {
		ret = append(ret, v)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].UpdatedAt < ret[j].UpdatedAt
	})

	return ret, nil
}

// GetUpdatedRetrievalRecordsSince returns all the retrieval records that got created or updated
// since sinceNano.
func (s *Store) GetUpdatedRetrievalRecordsSince(sinceNano int64, limit int) ([]deals.RetrievalDealRecord, error) {
	if limit < 0 {
		return nil, fmt.Errorf("limit is negative")
	}

	q := query.Query{
		Prefix: dsRetrievalUpdatedAtIdx.String(),
	}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing updated retrieval records query result: %s", err)
		}
	}()

	mrdrs := make(map[datastore.Key]deals.RetrievalDealRecord)
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("getting query result: %s", r.Error)
		}
		unixNanoKey := datastore.NewKey(r.Key).Namespaces()[2]
		unixNano, err := strconv.ParseInt(unixNanoKey, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing unix nano key: %s", err)
		}

		if unixNano < sinceNano {
			continue
		}

		recordKey := datastore.RawKey(string(r.Value))
		buf, err := s.ds.Get(recordKey)
		if err != nil {
			return nil, fmt.Errorf("get updated retrieval deal record from store: %s", err)
		}
		var rr deals.RetrievalDealRecord
		if err := json.Unmarshal(buf, &rr); err != nil {
			return nil, fmt.Errorf("unmarshaling updated retrieval deal record from store: %s", err)
		}

		mrdrs[recordKey] = rr

		limit--
		if limit == 0 {
			break
		}
	}

	ret := make([]deals.RetrievalDealRecord, 0, len(mrdrs))
	for _, v := range mrdrs {
		ret = append(ret, v)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].UpdatedAt < ret[j].UpdatedAt
	})

	return ret, nil
}

func makePendingDealKey(c cid.Cid) datastore.Key {
	return dsBaseStoragePending.ChildString(util.CidToString(c))
}

func makeFinalDealKey(c cid.Cid) datastore.Key {
	return dsBaseStorageFinal.ChildString(util.CidToString(c))
}

func makeRetrievalKey(rr deals.RetrievalDealRecord) datastore.Key {
	str := fmt.Sprintf("%v%v%v%v", rr.Time, rr.Addr, rr.DealInfo.Miner, util.CidToString(rr.DealInfo.RootCid))
	sum := md5.Sum([]byte(str))
	return dsBaseRetrieval.ChildString(fmt.Sprintf("%x", sum[:]))
}

func makeStorageUpdatedAtIndexKey(unixNano int64) datastore.Key {
	return dsStorageUpdatedAtIdx.ChildString(strconv.FormatInt(unixNano, 10))
}

func makeRetrievalUpdatedAtIndexKey(unixNano int64) datastore.Key {
	return dsRetrievalUpdatedAtIdx.ChildString(strconv.FormatInt(unixNano, 10))
}
