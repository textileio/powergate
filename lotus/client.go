package lotus

import (
	"context"
	"net/http"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/textileio/filecoin/lotus/jsonrpc"
	"github.com/textileio/filecoin/lotus/types"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	lotusSyncStatusInterval = time.Second * 10
	log                     = logging.Logger("deals")
)

type API struct {
	Internal struct {
		ClientStartDeal   func(ctx context.Context, data cid.Cid, addr string, miner string, price types.BigInt, blocksDuration uint64) (*cid.Cid, error)
		ClientImport      func(ctx context.Context, path string) (cid.Cid, error)
		ClientGetDealInfo func(context.Context, cid.Cid) (*types.DealInfo, error)
		ChainNotify       func(context.Context) (<-chan []*types.HeadChange, error)
		StateListMiners   func(context.Context, *types.TipSet) ([]string, error)
		ClientQueryAsk    func(ctx context.Context, p peer.ID, miner string) (*types.SignedStorageAsk, error)
		StateMinerPeerID  func(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error)
		Version           func(context.Context) (types.Version, error)
		SyncState         func(context.Context) (*types.SyncState, error)
	}
}

// New creates a new client to Lotus API
func New(addr string, authToken string) (*API, func(), error) {
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}
	var api API
	closer, err := jsonrpc.NewMergeClient("ws://"+addr+"/rpc/v0", "Filecoin",
		[]interface{}{
			&api.Internal,
		}, headers)
	if err != nil {
		return nil, nil, err
	}

	if err := view.Register(vHeight); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go monitorLotusSync(ctx, &api)

	return &api, func() {
		cancel()
		closer()
	}, nil

}

func (a *API) ClientStartDeal(ctx context.Context, data cid.Cid, addr string, miner string, price types.BigInt, blocksDuration uint64) (*cid.Cid, error) {
	return a.Internal.ClientStartDeal(ctx, data, addr, miner, price, blocksDuration)
}
func (a *API) ClientImport(ctx context.Context, path string) (cid.Cid, error) {
	return a.Internal.ClientImport(ctx, path)
}
func (a *API) ClientGetDealInfo(ctx context.Context, cid cid.Cid) (*types.DealInfo, error) {
	return a.Internal.ClientGetDealInfo(ctx, cid)
}
func (a *API) ChainNotify(ctx context.Context) (<-chan []*types.HeadChange, error) {
	return a.Internal.ChainNotify(ctx)
}
func (a *API) StateListMiners(ctx context.Context, tipset *types.TipSet) ([]string, error) {
	return a.Internal.StateListMiners(ctx, tipset)
}
func (a *API) ClientQueryAsk(ctx context.Context, p peer.ID, miner string) (*types.SignedStorageAsk, error) {
	return a.Internal.ClientQueryAsk(ctx, p, miner)
}
func (a *API) StateMinerPeerID(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error) {
	return a.Internal.StateMinerPeerID(ctx, m, ts)
}
func (a *API) Version(ctx context.Context) (types.Version, error) {
	return a.Internal.Version(ctx)
}
func (a *API) SyncState(ctx context.Context) (*types.SyncState, error) {
	return a.Internal.SyncState(ctx)
}

func monitorLotusSync(ctx context.Context, c *API) {
	refreshHeightMetric(c)
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing lotus sync monitor")
			return
		case <-time.After(lotusSyncStatusInterval):
			refreshHeightMetric(c)
		}
	}
}

func refreshHeightMetric(c *API) {
	var h uint64
	state, err := c.SyncState(context.Background())
	if err != nil {
		log.Errorf("error when getting lotus sync status: %s", err)
		return
	}
	for _, w := range state.ActiveSyncs {
		if w.Height > h {
			h = w.Height
		}
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(h)))
}
