package deals

import (
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

func TestPutPendingDeal(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}})
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "b", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c2}})
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "c", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c3}})
	require.Nil(t, err)
}

func TestGetPendingDeals(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	now := time.Now().Unix()

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "a", Time: now, Pending: true, DealInfo: DealInfo{ProposalCid: c1}})
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "b", Time: now + 1, Pending: true, DealInfo: DealInfo{ProposalCid: c2}})
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{Addr: "c", Time: now + 3, Pending: true, DealInfo: DealInfo{ProposalCid: c3}})
	require.Nil(t, err)

	res, err := s.getPendingDeals()
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, res[0].Addr, "a")
	require.Equal(t, res[1].Addr, "b")
	require.Equal(t, res[2].Addr, "c")
}

func TestDeletePendingDeal(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	dr := DealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}}
	err = s.putPendingDeal(dr)
	require.Nil(t, err)

	res, err := s.getPendingDeals()
	require.Nil(t, err)
	require.Len(t, res, 1)

	err = s.deletePendingDeal(dr.DealInfo.ProposalCid)
	require.Nil(t, err)

	res, err = s.getPendingDeals()
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestPutDealRecord(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	pdr := DealRecord{Addr: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}}
	err = s.putPendingDeal(pdr)
	require.Nil(t, err)

	res, err := s.getPendingDeals()
	require.Nil(t, err)
	require.Len(t, res, 1)

	dr := DealRecord{Addr: "a", Time: time.Now().Unix(), Pending: false, DealInfo: DealInfo{ProposalCid: c1}}

	err = s.putFinalDeal(dr)
	require.Nil(t, err)

	res, err = s.getPendingDeals()
	require.Nil(t, err)
	require.Empty(t, res)
}

func TestGetDealRecords(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	now := time.Now().Unix()

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	dr1 := DealRecord{Addr: "a", Time: now, Pending: false, DealInfo: DealInfo{ProposalCid: c1}}
	err = s.putFinalDeal(dr1)
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	dr2 := DealRecord{Addr: "b", Time: now + 1, Pending: false, DealInfo: DealInfo{ProposalCid: c2}}
	err = s.putFinalDeal(dr2)
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	dr3 := DealRecord{Addr: "c", Time: now + 2, Pending: false, DealInfo: DealInfo{ProposalCid: c3}}
	err = s.putFinalDeal(dr3)
	require.Nil(t, err)

	res, err := s.getFinalDeals()
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, res[0].Addr, "a")
	require.Equal(t, res[1].Addr, "b")
	require.Equal(t, res[2].Addr, "c")
}

func TestPutRetrievalRecords(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())
	now := time.Now().Unix()
	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	rr := RetrievalRecord{Time: now, Addr: "from", RetrievalInfo: RetrievalInfo{PieceCID: c1, Miner: "miner"}}
	err = s.putRetrieval(rr)
	require.Nil(t, err)
}

func TestGetRetrievalDeals(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())
	now := time.Now().Unix()

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	rr := RetrievalRecord{Time: now, Addr: "from", RetrievalInfo: RetrievalInfo{PieceCID: c1, Miner: "miner"}}
	err = s.putRetrieval(rr)
	require.Nil(t, err)

	require.Nil(t, err)
	rr = RetrievalRecord{Time: now + 1, Addr: "from", RetrievalInfo: RetrievalInfo{PieceCID: c1, Miner: "miner"}}
	err = s.putRetrieval(rr)
	require.Nil(t, err)

	require.Nil(t, err)
	rr = RetrievalRecord{Time: now + 2, Addr: "from", RetrievalInfo: RetrievalInfo{PieceCID: c1, Miner: "miner"}}
	err = s.putRetrieval(rr)
	require.Nil(t, err)

	res, err := s.getRetrievals()
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, res[0].Time, now)
	require.Equal(t, res[1].Time, now+1)
	require.Equal(t, res[2].Time, now+2)
}
