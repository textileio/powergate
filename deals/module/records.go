package module

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	datatransfer "github.com/filecoin-project/go-data-transfer"
	sm "github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/util"
)

var (
	errWatchingTimeout         = "pow watching timeout"
	errWatchingUnexpectedClose = "pow watching unexpected closing"
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

// GetUpdatedStorageDealRecordsSince returns all the storage deal records that got created or updated
// since sinceNano.
func (m *Module) GetUpdatedStorageDealRecordsSince(sinceNano int64, limit int) ([]deals.StorageDealRecord, error) {
	r, err := m.store.GetUpdatedStorageDealRecordsSince(sinceNano, limit)
	if err != nil {
		return nil, fmt.Errorf("get updated storage deal records from store: %s", err)
	}

	return r, nil
}

// GetUpdatedRetrievalRecordsSince returns all the retrieval records that got created or updated
// since sinceNano.
func (m *Module) GetUpdatedRetrievalRecordsSince(sinceNano int64, limit int) ([]deals.RetrievalDealRecord, error) {
	r, err := m.store.GetUpdatedRetrievalRecordsSince(sinceNano, limit)
	if err != nil {
		return nil, fmt.Errorf("get updated retrieval records from store: %s", err)
	}

	return r, nil
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

func (m *Module) recordDeal(params *api.StartDealParams, proposalCid cid.Cid, dataSize int64) {
	di := deals.StorageDealInfo{
		Duration:      params.MinBlocksDuration,
		PricePerEpoch: params.EpochPrice.Uint64(),
		Miner:         params.Miner.String(),
		ProposalCid:   proposalCid,
	}
	record := deals.StorageDealRecord{
		RootCid:      params.Data.Root,
		Addr:         params.Wallet.String(),
		Time:         time.Now().Unix(),
		DealInfo:     di,
		Pending:      true,
		TransferSize: dataSize,
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
		dr.ErrMsg = errMsg
		if err := m.store.PutStorageDeal(dr); err != nil {
			log.Errorf("erroring pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
		return
	}
	if info.State != sm.StorageDealActive {
		log.Infof("pending deal for proposal cid %s isn't active yet, erroring pending deal", util.CidToString(dr.DealInfo.ProposalCid))

		dr.ErrMsg = fmt.Sprintf("deal failed with status %s", sm.DealStates[info.State])
		if err := m.store.PutStorageDeal(dr); err != nil {
			log.Errorf("erroring pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
	} else {
		di, err := fromLotusDealInfo(ctx, lapi, info)
		if err != nil {
			dr.ErrMsg = fmt.Sprintf("converting proposal cid %s from lotus deal info: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
			log.Errorf(dr.ErrMsg)
			if err := m.store.PutStorageDeal(dr); err != nil {
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

	dealUpdates, err := m.Watch(ctx, dr.DealInfo.ProposalCid)
	if err != nil {
		log.Errorf("watching proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		return
	}
	dataTransferUpdates := make(chan dataTransferUpdate)
	go func() {
		if err = m.dataTransferUpdates(ctx, dr.DealInfo.ProposalCid, dataTransferUpdates); err != nil {
			log.Errorf("watching data transfer updates: %s", err)
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Infof("watching proposal cid %s timed out, erroring pending deal", util.CidToString(dr.DealInfo.ProposalCid))
			dr.ErrMsg = errWatchingTimeout
			if err := m.store.PutStorageDeal(dr); err != nil {
				log.Errorf("erroring pending deal: %v", err)
			}
			return
		case info, ok := <-dealUpdates:
			if !ok {
				log.Errorf("deal updates channel unexpectedly closed for proposal cid %s", util.CidToString(dr.DealInfo.ProposalCid))
				dr.ErrMsg = errWatchingUnexpectedClose
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("erroring pending deal: %v", err)
				}
				return
			}

			switch info.StateID {
			// Final status (`return`)
			case sm.StorageDealActive:
				dr.DealInfo = info
				dr.Pending = false
				if dr.SealingStart > 0 && dr.SealingEnd == 0 {
					dr.SealingEnd = time.Now().Unix()
				}
				log.Infof("proposal cid %s is active, storing deal record", util.CidToString(info.ProposalCid))
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("storing proposal cid %s deal record: %v", util.CidToString(info.ProposalCid), err)
				}
				return
			case sm.StorageDealProposalNotFound, sm.StorageDealProposalRejected, sm.StorageDealFailing, sm.StorageDealError:
				log.Infof("proposal cid %s failed with state %s, deleting pending deal", util.CidToString(info.ProposalCid), sm.DealStates[info.StateID])

				dr.ErrMsg = fmt.Sprintf("deal failed with status %s", sm.DealStates[info.StateID])
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("saving successful storage deal record: %v", err)
				}
				return

			// This case is just being paranoid of dataTransferUpdates not reporting the ending transfer time.
			// If we're in this status, data transfer should already finished, so just check if we got into
			// that situation and account for it. Shouldn't happen if Lotus doesn't misreports events.
			case sm.StorageDealCheckForAcceptance, sm.StorageDealProposalAccepted, sm.StorageDealAwaitingPreCommit:
				if dr.DataTransferStart > 0 && dr.DataTransferEnd == 0 {
					dr.DataTransferEnd = time.Now().Unix()
					if err := m.store.PutStorageDeal(dr); err != nil {
						log.Errorf("saving data transfer start time: %s", err)
					}
				}
			case sm.StorageDealSealing:
				if dr.SealingStart == 0 {
					dr.SealingStart = time.Now().Unix()
					if err := m.store.PutStorageDeal(dr); err != nil {
						log.Errorf("saving sealing start time: %s", err)
					}
				}
			}
		case u := <-dataTransferUpdates:
			if dr.DataTransferStart == 0 && !u.start.IsZero() {
				dr.DataTransferStart = u.start.Unix()
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("saving data transfer start time: %s", err)
				}
			}
			if dr.DataTransferEnd == 0 && !u.end.IsZero() {
				dr.DataTransferEnd = u.end.Unix()
				if err := m.store.PutStorageDeal(dr); err != nil {
					log.Errorf("saving data transfer end time: %s", err)
				}
			}
		}
	}
}

type dataTransferUpdate struct {
	start time.Time
	end   time.Time
}

type dealTransferVoucher struct {
	Proposal cid.Cid
}

func (m *Module) dataTransferUpdates(ctx context.Context, proposalCid cid.Cid, updates chan<- dataTransferUpdate) error {
	lapi, cls, err := m.clientBuilder(ctx)
	if err != nil {
		return fmt.Errorf("measure data transfer creating client: %s", err)
	}
	defer cls()
	ch, err := lapi.ClientDataTransferUpdates(ctx)
	if err != nil {
		return fmt.Errorf("measure data transfer getting update channel from lotus: %s", err)
	}

	var dtStart time.Time
	for {
		select {
		case <-ctx.Done():
			return nil
		case u, ok := <-ch:
			if !ok {
				return fmt.Errorf("data transfer updates channel unexpectedly closed: %s", err)
			}
			var dtVoucher dealTransferVoucher
			if err := json.Unmarshal([]byte(u.Voucher), &dtVoucher); err != nil {
				continue
			}
			if dtVoucher.Proposal != proposalCid {
				continue
			}
			switch u.Status {
			case datatransfer.Ongoing:
				if dtStart.IsZero() {
					dtStart = time.Now()
					updates <- dataTransferUpdate{start: dtStart, end: time.Time{}}
				}
			case datatransfer.Completed:
				updates <- dataTransferUpdate{start: dtStart, end: time.Now()}
				return nil
			case datatransfer.Failed, datatransfer.Cancelled:
				return nil
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
