package store

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

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

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("cid info not found")

	log = logging.Logger("deals-records")
)

type Store struct {
	ds   datastore.TxnDatastore
	lock sync.Mutex
}

func New(ds datastore.TxnDatastore) *Store {
	return &Store{
		ds: ds,
	}
}

func (s *Store) PutStorageDeal(dr deals.StorageDealRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return putStorageDeal(s.ds, dr)
}

func (s *Store) ErrorPendingDeal(dr deals.StorageDealRecord, errorMsg string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !dr.Pending {
		return fmt.Errorf("the deal record isn't pending")
	}

	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return fmt.Errorf("creating transaction: %s", err)
	}
	defer txn.Discard()

	if err := txn.Delete(makePendingDealKey(dr.DealInfo.ProposalCid)); err != nil {
		return fmt.Errorf("delete pending deal storage record: %s", err)
	}

	dr.ErrMsg = errorMsg
	dr.Pending = false

	if err := putStorageDeal(txn, dr); err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %s", err)
	}

	return nil
}

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

func (s *Store) PutRetrieval(rr deals.RetrievalDealRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	buf, err := json.Marshal(rr)
	if err != nil {
		return fmt.Errorf("marshaling RetrievalRecord: %s", err)
	}
	if err := s.ds.Put(makeRetrievalKey(rr), buf); err != nil {
		return fmt.Errorf("put RetrievalRecord: %s", err)
	}
	return nil
}

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

func putStorageDeal(dsw datastore.Write, dr deals.StorageDealRecord) error {
	buf, err := json.Marshal(dr)
	if err != nil {
		return fmt.Errorf("marshaling PendingDeal: %s", err)
	}

	var key datastore.Key
	if dr.Pending {
		key = makePendingDealKey(dr.DealInfo.ProposalCid)
	} else {
		key = makeFinalDealKey(dr.DealInfo.ProposalCid)
	}

	if err := dsw.Put(key, buf); err != nil {
		return fmt.Errorf("put storage deal record: %s", err)
	}

	return nil
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
