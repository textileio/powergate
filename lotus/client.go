package lotus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/lotus/jsonrpc"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/util"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	lotusSyncStatusInterval = time.Second * 10
	log                     = logging.Logger("deals")
)

type API struct {
	Internal struct {
		ClientStartDeal        func(ctx context.Context, data cid.Cid, addr string, miner string, price types.BigInt, blocksDuration uint64) (*cid.Cid, error)
		ClientImport           func(ctx context.Context, path string) (cid.Cid, error)
		ClientGetDealInfo      func(context.Context, cid.Cid) (*types.DealInfo, error)
		ChainNotify            func(context.Context) (<-chan []*types.HeadChange, error)
		StateListMiners        func(context.Context, *types.TipSet) ([]string, error)
		ClientQueryAsk         func(ctx context.Context, p peer.ID, miner string) (*types.SignedStorageAsk, error)
		StateMinerPeerID       func(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error)
		Version                func(context.Context) (types.Version, error)
		SyncState              func(context.Context) (*types.SyncState, error)
		WalletNew              func(context.Context, string) (string, error)
		WalletBalance          func(context.Context, string) (types.BigInt, error)
		StateMinerPower        func(context.Context, string, *types.TipSet) (types.MinerPower, error)
		ChainHead              func(context.Context) (*types.TipSet, error)
		ChainGetTipSet         func(context.Context, types.TipSetKey) (*types.TipSet, error)
		StateChangedActors     func(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
		ChainReadObj           func(context.Context, cid.Cid) ([]byte, error)
		StateReadState         func(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error)
		StateGetActor          func(ctx context.Context, actor string, ts *types.TipSet) (*types.Actor, error)
		ChainGetTipSetByHeight func(context.Context, uint64, *types.TipSet) (*types.TipSet, error)
		ChainGetPath           func(context.Context, types.TipSetKey, types.TipSetKey) ([]*types.HeadChange, error)
		ChainGetGenesis        func(context.Context) (*types.TipSet, error)
	}
}

// New creates a new client to Lotus API
func New(maddr ma.Multiaddr, authToken string) (*API, func(), error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, nil, err
	}
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}
	var api API
	closer, err := jsonrpc.NewMergeClient("ws://"+addr+"/rpc/v0", "Filecoin",
		[]interface{}{
			&api.Internal,
		}, headers, jsonrpc.WithReconnect(true, time.Second*3, 0))
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
	cid, err := a.Internal.ClientStartDeal(ctx, data, addr, miner, price, blocksDuration)
	if err != nil {
		return nil, fmt.Errorf("error when calling ClientStartDeal: %s", err)
	}
	return cid, nil
}
func (a *API) ClientImport(ctx context.Context, path string) (cid.Cid, error) {
	c, err := a.Internal.ClientImport(ctx, path)
	if err != nil {
		return cid.Undef, fmt.Errorf("error when calling ClientImport: %s", err)
	}
	return c, nil
}
func (a *API) ClientGetDealInfo(ctx context.Context, cid cid.Cid) (*types.DealInfo, error) {
	di, err := a.Internal.ClientGetDealInfo(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("error when calling ClientGetDealInfo: %s", err)
	}
	return di, err
}
func (a *API) ChainNotify(ctx context.Context) (<-chan []*types.HeadChange, error) {
	hc, err := a.Internal.ChainNotify(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainNotify: %s", err)
	}
	return hc, err
}
func (a *API) StateListMiners(ctx context.Context, tipset *types.TipSet) ([]string, error) {
	miners, err := a.Internal.StateListMiners(ctx, tipset)
	if err != nil {
		return nil, fmt.Errorf("error when calling StateListMiners: %s", err)
	}
	return miners, nil
}
func (a *API) ClientQueryAsk(ctx context.Context, p peer.ID, miner string) (*types.SignedStorageAsk, error) {
	asks, err := a.Internal.ClientQueryAsk(ctx, p, miner)
	if err != nil {
		return nil, fmt.Errorf("error when calling ClientQueryAsk: %s", err)
	}
	return asks, nil
}
func (a *API) StateMinerPeerID(ctx context.Context, m string, ts *types.TipSet) (peer.ID, error) {
	pid, err := a.Internal.StateMinerPeerID(ctx, m, ts)
	if err != nil {
		return peer.ID(""), fmt.Errorf("error when calling StateMinerPeerID: %s", err)
	}
	return pid, nil
}
func (a *API) Version(ctx context.Context) (types.Version, error) {
	v, err := a.Internal.Version(ctx)
	if err != nil {
		return types.Version{}, fmt.Errorf("error when calling Version: %s", err)
	}
	return v, nil
}
func (a *API) SyncState(ctx context.Context) (*types.SyncState, error) {
	ss, err := a.Internal.SyncState(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when calling SyncState: %s", err)
	}
	return ss, nil
}
func (a *API) WalletNew(ctx context.Context, typ string) (string, error) {
	addr, err := a.Internal.WalletNew(ctx, typ)
	if err != nil {
		return "", fmt.Errorf("error when calling WalletNew: %s", err)
	}
	return addr, nil
}
func (a *API) WalletBalance(ctx context.Context, addr string) (types.BigInt, error) {
	b, err := a.Internal.WalletBalance(ctx, addr)
	if err != nil {
		return types.NewInt(0), fmt.Errorf("error when calling WalletBalance: %s", err)
	}
	return b, nil
}
func (a *API) StateMinerPower(ctx context.Context, addr string, ts *types.TipSet) (types.MinerPower, error) {
	mp, err := a.Internal.StateMinerPower(ctx, addr, ts)
	if err != nil {
		return types.MinerPower{}, fmt.Errorf("error when calling StateMinerPower: %s", err)
	}
	return mp, nil
}
func (a *API) ChainHead(ctx context.Context) (*types.TipSet, error) {
	ts, err := a.Internal.ChainHead(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainHead: %s", err)
	}
	return ts, nil
}
func (a *API) ChainGetTipSet(ctx context.Context, tsk types.TipSetKey) (*types.TipSet, error) {
	ts, err := a.Internal.ChainGetTipSet(ctx, tsk)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainGetTipSet: %s", err)
	}
	return ts, nil
}
func (a *API) StateChangedActors(ctx context.Context, ocid cid.Cid, ncid cid.Cid) (map[string]types.Actor, error) {
	as, err := a.Internal.StateChangedActors(ctx, ocid, ncid)
	if err != nil {
		return nil, fmt.Errorf("error when calling StateChangedActors: %s", err)
	}
	return as, nil
}
func (a *API) ChainReadObj(ctx context.Context, cid cid.Cid) ([]byte, error) {
	state, err := a.Internal.ChainReadObj(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainReadObj: %s", err)
	}
	return state, nil
}
func (a *API) StateReadState(ctx context.Context, act *types.Actor, ts *types.TipSet) (*types.ActorState, error) {
	as, err := a.Internal.StateReadState(ctx, act, ts)
	if err != nil {
		return nil, fmt.Errorf("error when calling StateReadState: %s", err)
	}
	return as, nil
}
func (a *API) StateGetActor(ctx context.Context, actor string, ts *types.TipSet) (*types.Actor, error) {
	ac, err := a.Internal.StateGetActor(ctx, actor, ts)
	if err != nil {
		return nil, fmt.Errorf("error when calling StateGetActor: %s", err)
	}
	return ac, nil
}
func (a *API) ChainGetTipSetByHeight(ctx context.Context, height uint64, ts *types.TipSet) (*types.TipSet, error) {
	ts, err := a.Internal.ChainGetTipSetByHeight(ctx, height, ts)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainGetTipSetByHeight: %s", err)
	}
	return ts, nil
}
func (a *API) ChainGetPath(ctx context.Context, from types.TipSetKey, to types.TipSetKey) ([]*types.HeadChange, error) {
	tss, err := a.Internal.ChainGetPath(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainGetPath: %s", err)
	}
	return tss, nil
}
func (a *API) ChainGetGenesis(ctx context.Context) (*types.TipSet, error) {
	g, err := a.Internal.ChainGetGenesis(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when calling ChainGetGenesis: %s", err)
	}
	return g, nil
}

func monitorLotusSync(ctx context.Context, c *API) {
	refreshHeightMetric(ctx, c)
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing lotus sync monitor")
			return
		case <-time.After(lotusSyncStatusInterval):
			refreshHeightMetric(ctx, c)
		}
	}
}

func refreshHeightMetric(ctx context.Context, c *API) {
	heaviest, err := c.ChainHead(ctx)
	if err != nil {
		log.Errorf("error when getting lotus sync status: %s", err)
		return
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height)))
}
