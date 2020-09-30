package module

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/util"
)

const (
	chanWriteTimeout = time.Second
	dealTimeout      = time.Hour * 24

	defaultDealStartOffset = 72 * 60 * 60 / util.EpochDurationSeconds // 72hs
)

var (
	// ErrRetrievalNoAvailableProviders indicates that the data isn't available on any provided
	// to be retrieved.
	ErrRetrievalNoAvailableProviders = errors.New("no providers to retrieve the data")
	// ErrDealNotFound indicates a particular ProposalCid from a deal isn't found on-chain. Currently,
	// in Lotus this indicates that it may never existed on-chain, or it existed but it already expired
	// (currEpoch > StartEpoch+Duration).
	ErrDealNotFound = errors.New("deal not found on-chain")

	log = logging.Logger("deals")
)

// Module exposes storage and monitoring from the market.
type Module struct {
	clientBuilder lotus.ClientBuilder
	cfg           *deals.Config
	store         *store
	pollDuration  time.Duration
}

// New creates a new Module.
func New(ds datastore.TxnDatastore, clientBuilder lotus.ClientBuilder, pollDuration time.Duration, opts ...deals.Option) (*Module, error) {
	var cfg deals.Config
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, err
		}
	}
	m := &Module{
		clientBuilder: clientBuilder,
		cfg:           &cfg,
		store:         newStore(ds),
	}
	m.initPendingDeals()
	return m, nil
}

// Import imports raw data in the Filecoin client. The isCAR flag indicates if the data
// is already in CAR format, so it shouldn't be encoded into a UnixFS DAG in the Filecoin client.
// It returns the imported data cid and the data size.
func (m *Module) Import(ctx context.Context, data io.Reader, isCAR bool) (cid.Cid, int64, error) {
	f, err := ioutil.TempFile(m.cfg.ImportPath, "import-*")
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("error when creating tmpfile: %s", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("closing storing file: %s", err)
		}
	}()
	var size int64
	if size, err = io.Copy(f, data); err != nil {
		return cid.Undef, 0, fmt.Errorf("error when copying data to tmpfile: %s", err)
	}
	ref := api.FileRef{
		Path:  f.Name(),
		IsCAR: isCAR,
	}
	api, cls, err := m.clientBuilder()
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	res, err := api.ClientImport(ctx, ref)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("error when importing data: %s", err)
	}
	return res.Root, size, nil
}

// Store create Deal Proposals with all miners indicated in dcfgs. The epoch price
// is automatically calculated considering each miner epoch price and piece size.
// The data of dataCid should be already imported to the Filecoin Client or should be
// accessible to it. (e.g: is integrated with an IPFS node).
func (m *Module) Store(ctx context.Context, waddr string, dataCid cid.Cid, pieceSize uint64, dcfgs []deals.StorageDealConfig, minDuration uint64) ([]deals.StoreResult, error) {
	if minDuration < util.MinDealDuration {
		return nil, fmt.Errorf("duration %d should be greater or equal to %d", minDuration, util.MinDealDuration)
	}
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return nil, fmt.Errorf("parsing wallet address: %s", err)
	}
	lapi, cls, err := m.clientBuilder()
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	ts, err := lapi.ChainHead(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting chaing head: %s", err)
	}
	res := make([]deals.StoreResult, len(dcfgs))
	for i, c := range dcfgs {
		maddr, err := address.NewFromString(c.Miner)
		if err != nil {
			log.Errorf("invalid miner address %v: %s", c, err)
			res[i] = deals.StoreResult{
				Config: c,
			}
			continue
		}
		dealStartOffset := c.DealStartOffset
		if dealStartOffset == 0 {
			dealStartOffset = defaultDealStartOffset
		}
		params := &api.StartDealParams{
			Data: &storagemarket.DataRef{
				TransferType: storagemarket.TTGraphsync,
				Root:         dataCid,
			},
			MinBlocksDuration: minDuration,
			EpochPrice:        big.Div(big.Mul(big.NewIntUnsigned(c.EpochPrice), big.NewIntUnsigned(pieceSize)), abi.NewTokenAmount(1<<30)),
			Miner:             maddr,
			Wallet:            addr,
			FastRetrieval:     c.FastRetrieval,
			DealStartEpoch:    ts.Height() + abi.ChainEpoch(dealStartOffset),
		}
		p, err := lapi.ClientStartDeal(ctx, params)
		if err != nil {
			log.Errorf("starting deal with %v: %s", c, err)
			res[i] = deals.StoreResult{
				Config:  c,
				Message: err.Error(),
			}
			continue
		}
		res[i] = deals.StoreResult{
			Config:      c,
			ProposalCid: *p,
			Success:     true,
		}
		m.recordDeal(params, *p)
	}
	return res, nil
}

// Fetch fetches deal data to the underlying blockstore of the Filecoin client.
// This API is meant for clients that use external implementations of blockstores with
// their own API, e.g: IPFS.
func (m *Module) Fetch(ctx context.Context, waddr string, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string) (string, <-chan marketevents.RetrievalEvent, error) {
	lapi, cls, err := m.clientBuilder()
	if err != nil {
		return "", nil, fmt.Errorf("creating lotus client: %s", err)
	}

	miner, events, err := m.retrieve(ctx, lapi, cls, waddr, payloadCid, pieceCid, miners, nil)
	if err != nil {
		return "", nil, err
	}
	return miner, events, nil
}

// Retrieve retrieves Deal data. It returns the miner address where the data
// is being fetched from, and a byte reader to read the retrieved data.
func (m *Module) Retrieve(ctx context.Context, waddr string, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string, CAREncoding bool) (string, io.ReadCloser, error) {
	rf, err := ioutil.TempDir(m.cfg.ImportPath, "retrieve-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir for retrieval: %s", err)
	}
	ref := api.FileRef{
		Path:  filepath.Join(rf, "ret"),
		IsCAR: CAREncoding,
	}

	lapi, cls, err := m.clientBuilder()
	if err != nil {
		return "", nil, fmt.Errorf("creating lotus client: %s", err)
	}
	miner, events, err := m.retrieve(ctx, lapi, cls, waddr, payloadCid, pieceCid, miners, &ref)
	if err != nil {
		return "", nil, fmt.Errorf("retrieving from lotus: %s", err)
	}
	for e := range events {
		if e.Err != "" {
			return "", nil, fmt.Errorf("in progress retrieval error: %s", e.Err)
		}
	}
	f, err := os.Open(ref.Path)
	if err != nil {
		return "", nil, fmt.Errorf("opening retrieved file: %s", err)
	}

	return miner, &autodeleteFile{File: f}, nil
}

func (m *Module) retrieve(ctx context.Context, lapi *apistruct.FullNodeStruct, lapiCls func(), waddr string, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string, ref *api.FileRef) (string, <-chan marketevents.RetrievalEvent, error) {
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return "", nil, fmt.Errorf("parsing wallet address: %s", err)
	}

	// Ask each miner about costs and information about retrieving this data.
	var offers []api.QueryOffer
	for _, mi := range miners {
		a, err := address.NewFromString(mi)
		if err != nil {
			log.Infof("parsing miner address: %s", err)
		}
		qo, err := lapi.ClientMinerQueryOffer(ctx, a, payloadCid, pieceCid)
		if err != nil {
			log.Infof("asking miner %s query-offer failed: %s", m, err)
			continue
		}
		offers = append(offers, qo)
	}

	// If no miners available, fail.
	if len(offers) == 0 {
		return "", nil, ErrRetrievalNoAvailableProviders
	}

	// Sort received options by price.
	sort.Slice(offers, func(a, b int) bool { return offers[a].MinPrice.LessThan(offers[b].MinPrice) })

	out := make(chan marketevents.RetrievalEvent, 1)
	var events <-chan marketevents.RetrievalEvent

	// Try with sorted miners until we got in the process of receiving data.
	var o api.QueryOffer
	for _, o = range offers {
		events, err = lapi.ClientRetrieveWithEvents(ctx, o.Order(addr), ref)
		if err != nil {
			log.Infof("fetching/retrieving cid %s from %s: %s", payloadCid, o.Miner, err)
			continue
		}
		break
	}

	go func() {
		defer lapiCls()
		defer close(out)
		// Redirect received events to the output channel
		var errored, canceled bool
	Loop:
		for {
			select {
			case <-ctx.Done():
				log.Infof("in progress retrieval canceled")
				canceled = true
				break Loop
			case e, ok := <-events:
				if !ok {
					break Loop
				}
				if e.Err != "" {
					log.Infof("in progress retrieval errored: %s", err)
					errored = true
				}
				out <- e
			}
		}

		// Only register retrieval if successful
		if !errored && !canceled {
			m.recordRetrieval(waddr, o)
		}
	}()

	return o.Miner.String(), out, nil
}

// GetDealStatus returns the current status of the deal, and a flag indicating if the miner of the deal was slashed.
// If the deal doesn't exist, *or has expired* it will return ErrDealNotFound. There's not actual way of distinguishing
// both scenarios in Lotus.
func (m *Module) GetDealStatus(ctx context.Context, pcid cid.Cid) (storagemarket.StorageDealStatus, bool, error) {
	lapi, cls, err := m.clientBuilder()
	if err != nil {
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	di, err := lapi.ClientGetDealInfo(ctx, pcid)
	if err != nil {
		if strings.Contains(err.Error(), "datastore: key not found") {
			return storagemarket.StorageDealUnknown, false, ErrDealNotFound
		}
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("getting deal info: %s", err)
	}
	md, err := lapi.StateMarketStorageDeal(ctx, di.DealID, types.EmptyTSK)
	if err != nil {
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("get storage state: %s", err)
	}
	return di.State, md.State.SlashEpoch != -1, nil
}

// Watch returns a channel with state changes of indicated proposals.
func (m *Module) Watch(ctx context.Context, proposals []cid.Cid) (<-chan deals.StorageDealInfo, error) {
	if len(proposals) == 0 {
		return nil, fmt.Errorf("proposals list can't be empty")
	}
	ch := make(chan deals.StorageDealInfo)
	go func() {
		defer close(ch)
		currentState := make(map[cid.Cid]*api.DealInfo)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.pollDuration):
				api, cls, err := m.clientBuilder()
				if err != nil {
					log.Errorf("creating lotus client: %s", err)
					continue
				}
				if err := notifyChanges(ctx, api, currentState, proposals, ch); err != nil {
					log.Errorf("pushing new proposal states: %s", err)
				}
				cls()
			}
		}
	}()
	return ch, nil
}

// ListStorageDealRecords lists storage deals according to the provided options.
func (m *Module) ListStorageDealRecords(opts ...deals.ListDealRecordsOption) ([]deals.StorageDealRecord, error) {
	c := deals.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(&c)
	}

	if !c.IncludeFinal && !c.IncludePending {
		return nil, fmt.Errorf("you must specify one or both options of IncludePending and IncludeFinal")
	}

	var final []deals.StorageDealRecord
	if c.IncludeFinal {
		recs, err := m.store.getFinalDeals()
		if err != nil {
			return nil, fmt.Errorf("getting final deals: %v", err)
		}
		final = recs
	}

	var pending []deals.StorageDealRecord
	if c.IncludePending {
		recs, err := m.store.getPendingDeals()
		if err != nil {
			return nil, fmt.Errorf("getting pending deals: %v", err)
		}
		pending = recs
	}

	combined := append(final, pending...)

	var filtered []deals.StorageDealRecord

	if len(c.FromAddrs) > 0 || len(c.DataCids) > 0 {
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
			if includeViaFromAddrs && includeViaDataCids {
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
func (m *Module) ListRetrievalDealRecords(opts ...deals.ListDealRecordsOption) ([]deals.RetrievalDealRecord, error) {
	c := deals.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	ret, err := m.store.getRetrievals()
	if err != nil {
		return nil, fmt.Errorf("getting retrievals: %v", err)
	}

	var filtered []deals.RetrievalDealRecord

	if len(c.FromAddrs) > 0 || len(c.DataCids) > 0 {
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
			if includeViaFromAddrs && includeViaDataCids {
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

func (m *Module) initPendingDeals() {
	pendingDeals, err := m.store.getPendingDeals()
	if err != nil {
		log.Errorf("getting pending deals: %v", err)
		return
	}
	for _, dr := range pendingDeals {
		remaining := time.Until(time.Unix(dr.Time, 0).Add(dealTimeout))
		if remaining <= 0 {
			go m.finalizePendingDeal(dr)
		} else {
			go m.eventuallyFinalizeDeal(dr, remaining)
		}
	}
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
	if err := m.store.putPendingDeal(record); err != nil {
		log.Errorf("storing pending deal: %v", err)
		return
	}
	go m.eventuallyFinalizeDeal(record, dealTimeout)
}

func (m *Module) finalizePendingDeal(dr deals.StorageDealRecord) {
	lapi, cls, err := m.clientBuilder()
	if err != nil {
		log.Errorf("creating client: %s", err)
		return
	}
	defer cls()
	deletePending := func() {
		if err := m.store.deletePendingDeal(dr.DealInfo.ProposalCid); err != nil {
			log.Errorf("deleting pending deal for proposal cid %s: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	info, err := lapi.ClientGetDealInfo(ctx, dr.DealInfo.ProposalCid)
	if err != nil {
		log.Errorf("getting deal info: %v", err)
		deletePending()
		return
	}
	if info.State != storagemarket.StorageDealActive {
		log.Infof("pending deal for proposal cid %s isn't active yet, deleting pending deal", util.CidToString(dr.DealInfo.ProposalCid))
		deletePending()
	} else {
		di, err := fromLotusDealInfo(ctx, lapi, info)
		if err != nil {
			log.Errorf("converting proposal cid %s from lotus deal info: %v", util.CidToString(dr.DealInfo.ProposalCid), err)
			deletePending()
			return
		}
		record := deals.StorageDealRecord{
			RootCid:  dr.RootCid,
			Addr:     dr.Addr,
			Time:     time.Now().Unix(), // Note: This can be much later in time than the deal actually became active on chain
			DealInfo: di,
			Pending:  false,
		}
		if err := m.store.putFinalDeal(record); err != nil {
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
			log.Infof("watching proposal cid %s timed out, deleting pending deal", util.CidToString(dr.DealInfo.ProposalCid))
			if err := m.store.deletePendingDeal(dr.DealInfo.ProposalCid); err != nil {
				log.Errorf("deleting pending deal: %v", err)
			}
			return
		case info, ok := <-updates:
			if !ok {
				log.Errorf("updates channel unexpectedly closed for proposal cid %s", util.CidToString(dr.DealInfo.ProposalCid))
				if err := m.store.deletePendingDeal(dr.DealInfo.ProposalCid); err != nil {
					log.Errorf("deleting pending deal: %v", err)
				}
				return
			}
			if info.StateID == storagemarket.StorageDealActive {
				record := deals.StorageDealRecord{
					RootCid:  dr.RootCid,
					Addr:     dr.Addr,
					Time:     time.Now().Unix(),
					DealInfo: info,
					Pending:  false,
				}
				log.Infof("proposal cid %s is active, storing deal record", util.CidToString(info.ProposalCid))
				if err := m.store.putFinalDeal(record); err != nil {
					log.Errorf("storing proposal cid %s deal record: %v", util.CidToString(info.ProposalCid), err)
				}
				return
			} else if info.StateID == storagemarket.StorageDealProposalNotFound ||
				info.StateID == storagemarket.StorageDealProposalRejected ||
				info.StateID == storagemarket.StorageDealFailing {
				log.Infof("proposal cid %s failed with state %s, deleting pending deal", util.CidToString(info.ProposalCid), storagemarket.DealStates[info.StateID])
				if err := m.store.deletePendingDeal(info.ProposalCid); err != nil {
					log.Errorf("deleting pending deal: %v", err)
				}
				return
			}
		}
	}
}

func (m *Module) recordRetrieval(addr string, offer api.QueryOffer) {
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
	}
	if err := m.store.putRetrieval(rr); err != nil {
		log.Errorf("storing retrieval: %v", err)
	}
}

func notifyChanges(ctx context.Context, client *apistruct.FullNodeStruct, currState map[cid.Cid]*api.DealInfo, proposals []cid.Cid, ch chan<- deals.StorageDealInfo) error {
	for _, pcid := range proposals {
		dinfo, err := client.ClientGetDealInfo(ctx, pcid)
		if err != nil {
			return fmt.Errorf("getting deal proposal info %s: %s", pcid, err)
		}
		if currState[pcid] == nil || (*currState[pcid]).State != dinfo.State {
			currState[pcid] = dinfo
			newState, err := fromLotusDealInfo(ctx, client, dinfo)
			if err != nil {
				return fmt.Errorf("converting proposal cid %s from lotus deal info: %v", util.CidToString(pcid), err)
			}
			select {
			case <-ctx.Done():
				return nil
			case ch <- newState:
			case <-time.After(chanWriteTimeout):
				log.Warnf("dropping new state since chan is blocked")
			}
		}
	}
	return nil
}

func fromLotusDealInfo(ctx context.Context, client *apistruct.FullNodeStruct, dinfo *api.DealInfo) (deals.StorageDealInfo, error) {
	di := deals.StorageDealInfo{
		ProposalCid:   dinfo.ProposalCid,
		StateID:       dinfo.State,
		StateName:     storagemarket.DealStates[dinfo.State],
		Miner:         dinfo.Provider.String(),
		PieceCID:      dinfo.PieceCID,
		Size:          dinfo.Size,
		PricePerEpoch: dinfo.PricePerEpoch.Uint64(),
		Duration:      dinfo.Duration,
		DealID:        uint64(dinfo.DealID),
		Message:       dinfo.Message,
	}
	if dinfo.State == storagemarket.StorageDealActive {
		ocd, err := client.StateMarketStorageDeal(ctx, dinfo.DealID, types.EmptyTSK)
		if err != nil {
			return deals.StorageDealInfo{}, fmt.Errorf("getting on-chain deal info: %s", err)
		}
		di.ActivationEpoch = int64(ocd.State.SectorStartEpoch)
		di.StartEpoch = uint64(ocd.Proposal.StartEpoch)
	}
	return di, nil
}

type autodeleteFile struct {
	*os.File
}

func (af *autodeleteFile) Close() error {
	if err := af.File.Close(); err != nil {
		return fmt.Errorf("closing retrieval file: %s", err)
	}
	if err := os.Remove(af.File.Name()); err != nil {
		return fmt.Errorf("autodeleting retrieval file: %s", err)
	}
	return nil
}
