package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/tests"
	"github.com/textileio/powergate/v2/util"
)

func TestPutPendingDeal(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c1}})
	require.NoError(t, err)

	c2, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "b", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c2}})
	require.NoError(t, err)

	c3, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "c", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c3}})
	require.NoError(t, err)
}

func TestGetPendingDeals(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	now := time.Now().Unix()

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "a", Time: now, Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c1}})
	require.NoError(t, err)

	c2, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "b", Time: now + 1, Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c2}})
	require.NoError(t, err)

	c3, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.NoError(t, err)
	err = s.PutStorageDeal(deals.StorageDealRecord{Addr: "c", Time: now + 3, Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c3}})
	require.NoError(t, err)

	res, err := s.GetPendingStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 3)
	for _, sd := range res {
		require.Greater(t, sd.UpdatedAt, int64(0))
	}
}

func TestErrorPendingDeal(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	dr := deals.StorageDealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c1}}
	err = s.PutStorageDeal(dr)
	require.NoError(t, err)

	res, err := s.GetPendingStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 1)

	dr.Pending = false
	dr.ErrMsg = "ERROR TEST"
	err = s.PutStorageDeal(dr)
	require.NoError(t, err)

	res, err = s.GetPendingStorageDeals()
	require.NoError(t, err)
	require.Empty(t, res)

	res, err = s.GetFinalStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.False(t, res[0].Pending)
	require.Equal(t, "ERROR TEST", res[0].ErrMsg)
}

func TestPutDealRecord(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	pdr := deals.StorageDealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c1}}
	err = s.PutStorageDeal(pdr)
	require.NoError(t, err)

	res, err := s.GetPendingStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 1)
	res, err = s.GetFinalStorageDeals()
	require.NoError(t, err)
	require.Empty(t, res)

	dr := deals.StorageDealRecord{Addr: "a", Time: time.Now().Unix(), Pending: false, DealInfo: deals.StorageDealInfo{ProposalCid: c1}}

	err = s.PutStorageDeal(dr)
	require.NoError(t, err)

	res, err = s.GetPendingStorageDeals()
	require.NoError(t, err)
	require.Empty(t, res)
	res, err = s.GetFinalStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetDealRecords(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	now := time.Now().Unix()

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	dr1 := deals.StorageDealRecord{Addr: "a", Time: now, Pending: false, DealInfo: deals.StorageDealInfo{ProposalCid: c1}}
	err = s.PutStorageDeal(dr1)
	require.NoError(t, err)

	c2, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.NoError(t, err)
	dr2 := deals.StorageDealRecord{Addr: "b", Time: now + 1, Pending: false, DealInfo: deals.StorageDealInfo{ProposalCid: c2}}
	err = s.PutStorageDeal(dr2)
	require.NoError(t, err)

	c3, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.NoError(t, err)
	dr3 := deals.StorageDealRecord{Addr: "c", Time: now + 2, Pending: false, DealInfo: deals.StorageDealInfo{ProposalCid: c3}}
	err = s.PutStorageDeal(dr3)
	require.NoError(t, err)

	res, err := s.GetFinalStorageDeals()
	require.NoError(t, err)
	require.Len(t, res, 3)
}

func TestPutRetrievalRecords(t *testing.T) {
	s := New(tests.NewTxMapDatastore())
	now := time.Now().Unix()
	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	rr := deals.RetrievalDealRecord{Time: now, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)
}

func TestGetRetrievalDeals(t *testing.T) {
	s := New(tests.NewTxMapDatastore())
	now := time.Now().Unix()

	c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.NoError(t, err)
	rr := deals.RetrievalDealRecord{Time: now, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)

	rr = deals.RetrievalDealRecord{Time: now + 1, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)

	rr = deals.RetrievalDealRecord{Time: now + 2, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)

	res, err := s.GetRetrievals()
	require.NoError(t, err)
	require.Len(t, res, 3)
	for _, rr := range res {
		require.NotEmpty(t, rr.ID)
	}
}

func TestUpdatedAt(t *testing.T) {
	t.Run("retrieval", func(t *testing.T) {
		s := New(tests.NewTxMapDatastore())

		c1, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
		require.NoError(t, err)
		rr := deals.RetrievalDealRecord{Time: time.Now().Unix(), Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
		err = s.PutRetrieval(rr)
		require.NoError(t, err)

		res, err := s.GetRetrievals()
		require.NoError(t, err)
		updatedAt1 := res[0].UpdatedAt

		// Save again, which should update UpdatedAt
		err = s.PutRetrieval(rr)
		require.NoError(t, err)

		res, err = s.GetRetrievals()
		require.NoError(t, err)
		updatedAt2 := res[0].UpdatedAt

		require.Greater(t, updatedAt2, updatedAt1)
	})

	t.Run("storage-deal", func(t *testing.T) {
		s := New(tests.NewTxMapDatastore())

		c, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
		require.NoError(t, err)
		dr := deals.StorageDealRecord{Addr: "c", Time: time.Now().Unix(), Pending: false, DealInfo: deals.StorageDealInfo{ProposalCid: c}}
		err = s.PutStorageDeal(dr)
		require.NoError(t, err)

		res, err := s.GetFinalStorageDeals()
		require.NoError(t, err)
		updatedAt1 := res[0].UpdatedAt

		// Save again, which should update UpdatedAt
		err = s.PutStorageDeal(dr)
		require.NoError(t, err)

		res, err = s.GetFinalStorageDeals()
		require.NoError(t, err)
		updatedAt2 := res[0].UpdatedAt

		require.Greater(t, updatedAt2, updatedAt1)
	})
}

func TestStorageUpdatedSince(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	// Inception.
	t0 := time.Now().UnixNano()

	c1, _ := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	sr1 := deals.StorageDealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c1}}
	err := s.PutStorageDeal(sr1)
	require.NoError(t, err)

	// Checkpoint 1
	t1 := time.Now().UnixNano()

	c2, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.NoError(t, err)
	sr2 := deals.StorageDealRecord{Addr: "b", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c2}}
	err = s.PutStorageDeal(sr2)
	require.NoError(t, err)

	c3, err := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.NoError(t, err)
	sr3 := deals.StorageDealRecord{Addr: "c", Time: time.Now().Unix(), Pending: true, DealInfo: deals.StorageDealInfo{ProposalCid: c3}}
	err = s.PutStorageDeal(sr3)
	require.NoError(t, err)

	// Checkpoint 2
	t2 := time.Now().UnixNano()

	sr2Updated := sr2
	sr2Updated.Addr = "c2"
	err = s.PutStorageDeal(sr2Updated)
	require.NoError(t, err)

	// Checkpoint 3
	t3 := time.Now().UnixNano()

	// ## Start verifying changes.

	// Since t0, we should get 3 results, even if there were 4 changes (since sr2 as touched two times; we get the latest state).
	udr, err := s.GetUpdatedStorageDealRecordsSince(t0, 10)
	require.NoError(t, err)
	require.Len(t, udr, 3)
	require.Equal(t, sr1.Addr, udr[0].Addr)
	require.Equal(t, sr3.Addr, udr[1].Addr)
	require.Equal(t, sr2Updated.Addr, udr[2].Addr)

	// Since t1, we should get 2 results: sr3 and sr2Updated
	udr, err = s.GetUpdatedStorageDealRecordsSince(t1, 10)
	require.NoError(t, err)
	require.Len(t, udr, 2)
	require.Equal(t, sr3.Addr, udr[0].Addr)
	require.Equal(t, sr2Updated.Addr, udr[1].Addr)

	// Since t2, we should get 1 result: sr2Updated
	udr, err = s.GetUpdatedStorageDealRecordsSince(t2, 10)
	require.NoError(t, err)
	require.Len(t, udr, 1)
	require.Equal(t, sr2Updated.Addr, udr[0].Addr)

	// Since t3, we should get 0 results
	udr, err = s.GetUpdatedStorageDealRecordsSince(t3, 10)
	require.NoError(t, err)
	require.Len(t, udr, 0)

	// Test limit
	udr, err = s.GetUpdatedStorageDealRecordsSince(t0, 1)
	require.NoError(t, err)
	require.Len(t, udr, 1)
}

func TestRetrievalUpdatedSince(t *testing.T) {
	s := New(tests.NewTxMapDatastore())

	// Inception.
	t0 := time.Now().UnixNano()

	c1, _ := util.CidFromString("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	rr1 := deals.RetrievalDealRecord{Time: time.Now().UnixNano(), Addr: "c1", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err := s.PutRetrieval(rr1)
	require.NoError(t, err)

	// Checkpoint 1
	t1 := time.Now().UnixNano()

	rr2 := deals.RetrievalDealRecord{Time: time.Now().UnixNano(), Addr: "c2", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr2)
	require.NoError(t, err)

	rr3 := deals.RetrievalDealRecord{Time: time.Now().UnixNano(), Addr: "c3", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr3)
	require.NoError(t, err)

	// Checkpoint 2
	t2 := time.Now().UnixNano()

	rr2Updated := rr2
	rr2Updated.Addr = "c2"
	err = s.PutRetrieval(rr2Updated)
	require.NoError(t, err)

	// Checkpoint 3
	t3 := time.Now().UnixNano()

	// ## Start verifying changes.

	// Since t0, we should get 3 results, even if there were 4 changes (since rr2 as touched two times; we get the latest state).
	udr, err := s.GetUpdatedRetrievalRecordsSince(t0, 10)
	require.NoError(t, err)
	require.Len(t, udr, 3)
	require.Equal(t, rr1.Addr, udr[0].Addr)
	require.Equal(t, rr3.Addr, udr[1].Addr)
	require.Equal(t, rr2Updated.Addr, udr[2].Addr)

	// Since t1, we should get 2 results: sr3 and rr2Updated
	udr, err = s.GetUpdatedRetrievalRecordsSince(t1, 10)
	require.NoError(t, err)
	require.Len(t, udr, 2)
	require.Equal(t, rr3.Addr, udr[0].Addr)
	require.Equal(t, rr2Updated.Addr, udr[1].Addr)

	// Since t2, we should get 1 result: rr2Updated
	udr, err = s.GetUpdatedRetrievalRecordsSince(t2, 10)
	require.NoError(t, err)
	require.Len(t, udr, 1)
	require.Equal(t, rr2Updated.Addr, udr[0].Addr)

	// Since t3, we should get 0 results
	udr, err = s.GetUpdatedRetrievalRecordsSince(t3, 10)
	require.NoError(t, err)
	require.Len(t, udr, 0)

	// Test limit
	udr, err = s.GetUpdatedRetrievalRecordsSince(t0, 1)
	require.NoError(t, err)
	require.Len(t, udr, 1)
}
