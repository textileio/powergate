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
	err = s.putPendingDeal(DealRecord{From: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}})
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{From: "b", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c2}})
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{From: "c", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c3}})
	require.Nil(t, err)
}

func TestGetPendingDeals(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	now := time.Now().Unix()

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{From: "a", Time: now, Pending: true, DealInfo: DealInfo{ProposalCid: c1}})
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{From: "b", Time: now + 1, Pending: true, DealInfo: DealInfo{ProposalCid: c2}})
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	err = s.putPendingDeal(DealRecord{From: "c", Time: now + 3, Pending: true, DealInfo: DealInfo{ProposalCid: c3}})
	require.Nil(t, err)

	res, err := s.getPendingDeals()
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, res[0].From, "a")
	require.Equal(t, res[1].From, "b")
	require.Equal(t, res[2].From, "c")
}

func TestDeletePendingDeal(t *testing.T) {
	s := newStore(tests.NewTxMapDatastore())

	c1, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2D")
	require.Nil(t, err)
	dr := DealRecord{From: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}}
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
	pdr := DealRecord{From: "a", Time: time.Now().Unix(), Pending: true, DealInfo: DealInfo{ProposalCid: c1}}
	err = s.putPendingDeal(pdr)
	require.Nil(t, err)

	res, err := s.getPendingDeals()
	require.Nil(t, err)
	require.Len(t, res, 1)

	dr := DealRecord{From: "a", Time: time.Now().Unix(), Pending: false, DealInfo: DealInfo{ProposalCid: c1}}

	err = s.putDealRecord(dr)
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
	dr1 := DealRecord{From: "a", Time: now, Pending: false, DealInfo: DealInfo{ProposalCid: c1}}
	err = s.putDealRecord(dr1)
	require.Nil(t, err)

	c2, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2E")
	require.Nil(t, err)
	dr2 := DealRecord{From: "b", Time: now + 1, Pending: false, DealInfo: DealInfo{ProposalCid: c2}}
	err = s.putDealRecord(dr2)
	require.Nil(t, err)

	c3, err := cid.Decode("QmSnuWmxptJZdLJpKRarxBMS2Ju2oANVrgbr2xWbie9b2F")
	require.Nil(t, err)
	dr3 := DealRecord{From: "c", Time: now + 2, Pending: false, DealInfo: DealInfo{ProposalCid: c3}}
	err = s.putDealRecord(dr3)
	require.Nil(t, err)

	res, err := s.getDealRecords()
	require.Nil(t, err)
	require.Len(t, res, 3)
	require.Equal(t, res[0].From, "a")
	require.Equal(t, res[1].From, "b")
	require.Equal(t, res[2].From, "c")
}
