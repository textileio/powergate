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

	require.NoError(t, err)
	rr = deals.RetrievalDealRecord{Time: now + 1, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)

	require.NoError(t, err)
	rr = deals.RetrievalDealRecord{Time: now + 2, Addr: "from", DealInfo: deals.RetrievalDealInfo{RootCid: c1, Miner: "miner"}}
	err = s.PutRetrieval(rr)
	require.NoError(t, err)

	res, err := s.GetRetrievals()
	require.NoError(t, err)
	require.Len(t, res, 3)
}
