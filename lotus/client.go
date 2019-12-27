package lotus

import (
	"context"
	"net/http"

	"github.com/ipfs/go-cid"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/textileio/filecoin/lotus/jsonrpc"
	"github.com/textileio/filecoin/lotus/types"
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

	return &api, closer, err
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
