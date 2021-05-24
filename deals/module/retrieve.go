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
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/lotus/api"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/ipfs/go-cid"
)

var (
	// ErrRetrievalNoAvailableProviders indicates that the data isn't available on any provided
	// to be retrieved.
	ErrRetrievalNoAvailableProviders = errors.New("no providers to retrieve the data")
)

// Fetch fetches deal data to the underlying blockstore of the Filecoin client.
// This API is meant for clients that use external implementations of blockstores with
// their own API, e.g: IPFS.
func (m *Module) Fetch(ctx context.Context, waddr string, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string) (string, <-chan marketevents.RetrievalEvent, error) {
	lapi, cls, err := m.clientBuilder(ctx)
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

	lapi, cls, err := m.clientBuilder(ctx)
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

func (m *Module) retrieve(ctx context.Context, lapi *api.FullNodeStruct, lapiCls func(), waddr string, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string, ref *api.FileRef) (string, <-chan marketevents.RetrievalEvent, error) {
	addr, err := address.NewFromString(waddr)
	if err != nil {
		return "", nil, fmt.Errorf("parsing wallet address: %s", err)
	}

	sortedOffers := getRetrievalOffers(ctx, lapi, payloadCid, pieceCid, miners)
	if len(sortedOffers) == 0 {
		return "", nil, ErrRetrievalNoAvailableProviders
	}

	var events <-chan marketevents.RetrievalEvent

	// Try to make the retrieval in the specified offers order, until we
	// find one accepting providing the data.
	var o api.QueryOffer
	for _, o = range sortedOffers {
		events, err = lapi.ClientRetrieveWithEvents(ctx, o.Order(addr), ref)
		if err != nil {
			log.Infof("fetching/retrieving cid %s from %s: %s", payloadCid, o.Miner, err)
			continue
		}
		break
	}

	out := make(chan marketevents.RetrievalEvent, 1)
	go func() {
		defer lapiCls()
		defer close(out)
		m.metricRetrievalTracking.Add(ctx, 1)
		defer m.metricRetrievalTracking.Add(ctx, -1)

		// Redirect received events to the output channel
		var (
			canceled       bool
			errMsg         string
			dtStart, dtEnd time.Time
		)
		retrievalStartTime := time.Now()
		var bytesReceived uint64
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
					log.Infof("in progress retrieval errored: %s", e.Err)
					errMsg = e.Err
				}
				if dtStart.IsZero() && e.Event == retrievalmarket.ClientEventBlocksReceived {
					dtStart = time.Now()
				}
				if e.Event == retrievalmarket.ClientEventAllBlocksReceived {
					dtEnd = time.Now()
				}
				bytesReceived = e.BytesReceived
				out <- e
			}
		}

		if !canceled {
			// This is a fallback if for some reason the
			// expected event that signals the first block
			// transfer is missed or not received.
			// We fallback to the starting time of the retrieval,
			// which means that will account for possibly the
			// payment channel creation. This isn't ideal, but
			// it's better than missing the data.
			// We WARN just to signal this might be happening.
			if dtStart.IsZero() && errMsg == "" {
				dtStart = retrievalStartTime
				log.Warnf("retrieval data-transfer start fallback to retrieval start")
			}
			// This is a fallback to not receiving an expected
			// event in the retrieval. We just fallback to Now(),
			// which should always be pretty close to the real
			// event. We WARN just to signal this is happening.
			if dtEnd.IsZero() && errMsg == "" {
				dtEnd = time.Now()
				log.Warnf("retrieval data-transfer end fallback to retrieval end")
			}
			m.recordRetrieval(waddr, o, bytesReceived, dtStart, dtEnd, errMsg)
		}
	}()

	return o.MinerPeer.Address.String(), out, nil
}

func getRetrievalOffers(ctx context.Context, lapi *api.FullNodeStruct, payloadCid cid.Cid, pieceCid *cid.Cid, miners []string) []api.QueryOffer {
	// Ask each miner about costs and information about retrieving this data.
	var offers []api.QueryOffer
	for _, mi := range miners {
		a, err := address.NewFromString(mi)
		if err != nil {
			log.Infof("parsing miner address: %s", err)
		}
		qo, err := lapi.ClientMinerQueryOffer(ctx, a, payloadCid, pieceCid)
		if err != nil {
			log.Infof("asking miner %s query-offer failed: %s", a, err)
			continue
		}
		offers = append(offers, qo)
	}

	// Sort received options by price.
	sort.Slice(offers, func(a, b int) bool { return offers[a].MinPrice.LessThan(offers[b].MinPrice) })

	return offers
}
