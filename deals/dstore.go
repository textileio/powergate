package deals

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

var (
	dsBaseDeal    = datastore.NewKey("deal")
	dsBasePending = datastore.NewKey("pending")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("cid info not found")
)

type store struct {
	ds   datastore.Datastore
	lock sync.Mutex
}

func newStore(ds datastore.Datastore) *store {
	return &store{
		ds: ds,
	}
}

func (s *store) putPendingDeal(pd PendingDeal) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	buf, err := json.Marshal(pd)
	if err != nil {
		return fmt.Errorf("marshaling PendingDeal: %s", err)
	}
	if err := s.ds.Put(makePendingDealKey(pd.ProposalCid), buf); err != nil {
		return fmt.Errorf("put PendingDeal: %s", err)
	}
	return nil
}

func (s *store) deletePendingDeal(proposalCid cid.Cid) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.ds.Delete(makePendingDealKey(proposalCid)); err != nil {
		return fmt.Errorf("delete PendingDeal: %s", err)
	}
	return nil
}

func (s *store) getPendingDeals() ([]PendingDeal, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	q := query.Query{Prefix: dsBasePending.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []PendingDeal
	for r := range res.Next() {
		var pendingDeal PendingDeal
		if err := json.Unmarshal(r.Value, &pendingDeal); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, pendingDeal)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Time < ret[j].Time })
	return ret, nil
}

func (s *store) putDealRecord(dr DealRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.ds.Delete(makePendingDealKey(dr.DealInfo.ProposalCid))
	if err != nil {
		return fmt.Errorf("deleting PendingDeal: %s", err)
	}
	buf, err := json.Marshal(dr)
	if err != nil {
		return fmt.Errorf("marshaling DealRecord: %s", err)
	}
	if err := s.ds.Put(makeDealRecordKey(dr.DealInfo.ProposalCid), buf); err != nil {
		return fmt.Errorf("put DealRecord: %s", err)
	}
	return nil
}

func (s *store) getDealRecords() ([]DealRecord, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	time.Now().UnixNano()
	q := query.Query{Prefix: dsBaseDeal.String()}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []DealRecord
	for r := range res.Next() {
		var dealRecord DealRecord
		if err := json.Unmarshal(r.Value, &dealRecord); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, dealRecord)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Time < ret[j].Time })
	return ret, nil
}

func (s *store) close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.ds.Close()
	return err
}

func makePendingDealKey(c cid.Cid) datastore.Key {
	return dsBasePending.ChildString(c.String())
}

func makeDealRecordKey(c cid.Cid) datastore.Key {
	return dsBaseDeal.ChildString(c.String())
}
