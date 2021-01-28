package module

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/util"
)

var (
	errWatchingTimeout         = "pow: watching timeout"
	errWatchingUnexpectedClose = "pow: watching unexpected closing"
)

// ListStorageDealRecords lists storage deals according to the provided options.
func (m *Module) ListStorageDealRecords(opts ...deals.DealRecordsOption) ([]deals.StorageDealRecord, error) {
	c := deals.DealRecordsConfig{}
	for _, opt := range opts {
		opt(&c)
	}

	if !c.IncludeFinal && !c.IncludePending {
		return nil, fmt.Errorf("you must specify one or both options of IncludePending and IncludeFinal")
	}

	var final []deals.StorageDealRecord
	if c.IncludeFinal {
		recs, err := m.store.GetFinalStorageDeals()
		if err != nil {
			return nil, fmt.Errorf("getting final deals: %v", err)
		}
		final = recs
	}

	var pending []deals.StorageDealRecord
	if c.IncludePending {
		recs, err := m.store.GetPendingStorageDeals()
		if err != nil {
			return nil, fmt.Errorf("getting pending deals: %v", err)
		}
		pending = recs
	}

	combined := append(final, pending...)

	var filtered []deals.StorageDealRecord

	if len(c.FromAddrs) > 0 || len(c.DataCids) > 0 || !c.IncludeFailed {
		fromAddrsFilter := make(map[string]struct{})
		dataCidsFilter := make(map[string]struct{})
		for _, addr := range c.FromAddrs {
			fromAddrsFilter[addr] = struct{}{}
		}
		for _, cid := range c.DataCids {
			dataCidsFilter[cid] = struct{}{}
		}
		for _, record := range combined {
			_, inFromAddrsFilter := fromAddrsFilter[record.Addr]
			_, inDataCidsFilter := dataCidsFilter[util.CidToString(record.RootCid)]
			includeViaFromAddrs := len(c.FromAddrs) == 0 || inFromAddrsFilter
			includeViaDataCids := len(c.DataCids) == 0 || inDataCidsFilter
			includeViaIncludeFailed := !c.IncludeFailed || record.ErrMsg != ""
			if includeViaFromAddrs && includeViaDataCids && includeViaIncludeFailed {
				filtered = append(filtered, record)
			}
		}
	} else {
		filtered = combined
	}

	sort.Slice(filtered, func(i, j int) bool {
		l := filtered[j]
		r := filtered[i]
		if c.Ascending {
			l = filtered[i]
			r = filtered[j]
		}
		return l.Time < r.Time
	})

	return filtered, nil
}

// ListRetrievalDealRecords returns a list of retrieval deals according to the provided options.
func (m *Module) ListRetrievalDealRecords(opts ...deals.DealRecordsOption) ([]deals.RetrievalDealRecord, error) {
	c := deals.DealRecordsConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	ret, err := m.store.GetRetrievals()
	if err != nil {
		return nil, fmt.Errorf("getting retrievals: %v", err)
	}

	var filtered []deals.RetrievalDealRecord

	if len(c.FromAddrs) > 0 || len(c.DataCids) > 0 || !c.IncludeFailed {
		fromAddrsFilter := make(map[string]struct{})
		dataCidsFilter := make(map[string]struct{})
		for _, addr := range c.FromAddrs {
			fromAddrsFilter[addr] = struct{}{}
		}
		for _, cid := range c.DataCids {
			dataCidsFilter[cid] = struct{}{}
		}

		for _, record := range ret {
			_, inFromAddrsFilter := fromAddrsFilter[record.Addr]
			_, inDataCidsFilter := dataCidsFilter[util.CidToString(record.DealInfo.RootCid)]
			includeViaFromAddrs := len(c.FromAddrs) == 0 || inFromAddrsFilter
			includeViaDataCids := len(dataCidsFilter) == 0 || inDataCidsFilter
			includeViaIncludeFailed := !c.IncludeFailed || record.ErrMsg != ""
			if includeViaFromAddrs && includeViaDataCids && includeViaIncludeFailed {
				filtered = append(filtered, record)
			}
		}
	} else {
		filtered = ret
	}

	sort.Slice(filtered, func(i, j int) bool {
		l := filtered[j]
		r := filtered[i]
		if c.Ascending {
			l = filtered[i]
			r = filtered[j]
		}
		return l.Time < r.Time
	})

	return filtered, nil
}

func (m *Module) resumeWatchingPendingRecords() error {
	pendingStorageRecords, err := m.store.GetPendingStorageDeals()
	if err != nil {
		return fmt.Errorf("getting pending storage records from store: %v", err)
	}
	for _, dr := range pendingStorageRecords {
		remaining := time.Until(time.Unix(dr.Time, 0).Add(m.dealFinalityTimeout))
		if remaining <= 0 {
			go m.finalizePendingDeal(dr)
		} else {
			go m.eventuallyFinalizeDeal(dr, remaining)
		}
	}
	log.Infof("resumed watching %d pending storage deal records", len(pendingStorageRecords))

	return nil
}

func (m *Module) recordDeal(params *api.StartDealParams, proposalCid cid.Cid) {
	di := deals.StorageDealInfo{
		Duration:      params.MinBlocksDuration,
		PricePerEpoch: params.EpochPrice.Uint64(),
		Miner:         params.Miner.String(),
		ProposalCid:   proposalCid,
	}
	record := deals.StorageDealRecord{
		RootCid:  params.Data.Root,
		Addr:     params.Wallet.String(),
		Time:     time.Now().Unix(),
		DealInfo: di,
		Pending:  true,
	}
	log.Infof("storing pending deal record for proposal cid: %s", util.CidToString(proposalCid))
	if err := m.store.PutStorageDeal(record); err != nil {
		log.Errorf("storing pending deal: %v", err)
		return
	}
	go m.eventuallyFinalizeDeal(record, m.dealFinalityTimeout)
}

func (m *Module) finalizePendingDeal(dr deals.StorageDealRecord) {
	lapi, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		log.Errorf("finalize pending deal, creating client: %s", err)
		return
	}
	defer cls()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	info, err := robustClientGetDealInfo(ctx, lapi, dr.DealInfo.ProposalCid)
	if err != nil {
		errMsg := fmt.Sprintf("getting deal info: %s", err)
		log.Error(errMsg)
		if err := m.store.ErrorPendingDeal(dr, errMsg); err != nil {
			log.Errorf("erroring pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
		return
	}
	if info.State != storagemarket.StorageDealActive {
		log.Infof("pending deal for proposal cid %s isn't active yet, erroring pending deal", util.CidToString(dr.DealInfo.ProposalCid))

		errMsg := fmt.Sprintf("deal failed with status %s", storagemarket.DealStates[info.State])
		if err := m.store.ErrorPendingDeal(dr, errMsg); err != nil {
			log.Errorf("erroring pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
	} else {
		di, err := fromLotusDealInfo(ctx, lapi, info)
		if err != nil {
			errMsg := fmt.Sprintf("converting proposal cid %s from lotus deal info: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
			log.Errorf(errMsg)
			if err := m.store.ErrorPendingDeal(dr, errMsg); err != nil {
				log.Errorf("erroring pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
			}
			return
		}
		record := deals.StorageDealRecord{
			RootCid:  dr.RootCid,
			Addr:     dr.Addr,
			Time:     time.Now().Unix(), // Note: This can be much later in time than the deal actually became active on chain
			DealInfo: di,
			Pending:  false,
		}
		if err := m.store.PutStorageDeal(record); err != nil {
			log.Errorf("storing proposal cid %s deal record: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
	}
}

func (m *Module) eventuallyFinalizeDeal(dr deals.StorageDealRecord, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	updates, err := m.Watch(ctx, []cid.Cid{dr.DealInfo.ProposalCid})
	if err != nil {
		log.Errorf("watching proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			log.Infof("watching proposal cid %s timed out, erroring pending deal", util.CidToString(dr.DealInfo.ProposalCid))
			if err := m.store.ErrorPendingDeal(dr, errWatchingTimeout); err != nil {
				log.Errorf("erroring pending deal: %v", err)
			}
			return
		case info, ok := <-updates:
			if !ok {
				log.Errorf("updates channel unexpectedly closed for proposal cid %s", util.CidToString(dr.DealInfo.ProposalCid))
				if err := m.store.ErrorPendingDeal(dr, errWatchingUnexpectedClose); err != nil {
					log.Errorf("erroring pending deal: %v", err)
				}
				return
			}
			if info.StateID == storagemarket.StorageDealActive {
				dr.DealInfo = info
				dr.Pending = false
				log.Infof("proposal cid %s is active, storing deal record", util.CidToString(info.ProposalCid))
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("storing proposal cid %s deal record: %v", util.CidToString(info.ProposalCid), err)
				}
				return
			} else if info.StateID == storagemarket.StorageDealProposalNotFound ||
				info.StateID == storagemarket.StorageDealProposalRejected ||
				info.StateID == storagemarket.StorageDealFailing ||
				info.StateID == storagemarket.StorageDealError {
				log.Infof("proposal cid %s failed with state %s, deleting pending deal", util.CidToString(info.ProposalCid), storagemarket.DealStates[info.StateID])
				errMsg := fmt.Sprintf("deal failed with status %s", storagemarket.DealStates[info.StateID])
				if err := m.store.ErrorPendingDeal(dr, errMsg); err != nil {
					log.Errorf("erroring pending deal: %v", err)
				}
				return
			}
		}
	}
}

func (m *Module) recordRetrieval(addr string, offer api.QueryOffer, dtStart, dtEnd time.Time, errMsg string) {
	rr := deals.RetrievalDealRecord{
		Addr: addr,
		Time: time.Now().Unix(),
		DealInfo: deals.RetrievalDealInfo{
			RootCid:                 offer.Root,
			Size:                    offer.Size,
			MinPrice:                offer.MinPrice.Uint64(),
			Miner:                   offer.Miner.String(),
			MinerPeerID:             offer.MinerPeer.ID.String(),
			PaymentInterval:         offer.PaymentInterval,
			PaymentIntervalIncrease: offer.PaymentIntervalIncrease,
		},
		DataTransferStart: dtStart.Unix(),
		DataTransferEnd:   dtEnd.Unix(),
		ErrMsg:            errMsg,
	}
	if err := m.store.PutRetrieval(rr); err != nil {
		log.Errorf("storing retrieval: %v", err)
	}
}
