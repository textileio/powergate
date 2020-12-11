package module

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/ipfs/go-cid"
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
