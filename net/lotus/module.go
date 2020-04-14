package lotus

import (
	"context"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/lotus-client/api/apistruct"
	"github.com/textileio/powergate/net"
)

var _ net.Module = (*Module)(nil)

// Module exposes the filecoin wallet api.
type Module struct {
	api *apistruct.FullNodeStruct
}

// New creates a new net module
func New(api *apistruct.FullNodeStruct) *Module {
	m := &Module{
		api: api,
	}
	return m
}

func (m *Module) ListenAddr(ctx context.Context) (peer.AddrInfo, error) {
	return m.api.NetAddrsListen(ctx)
}

func (m *Module) ConnectPeer(ctx context.Context, addrInfo peer.AddrInfo) error {
	return m.api.NetConnect(ctx, addrInfo)
}

func (m *Module) DisconnectPeer(ctx context.Context, peerID peer.ID) error {
	return m.api.NetDisconnect(ctx, peerID)
}

func (m *Module) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return m.api.NetFindPeer(ctx, peerID)
}

func (m *Module) Peers(ctx context.Context) ([]peer.AddrInfo, error) {
	return m.api.NetPeers(ctx)
}

func (m *Module) Connectedness(ctx context.Context, peerID peer.ID) (net.Connectedness, error) {
	con, err := m.api.NetConnectedness(ctx, peerID)
	if err != nil {
		return net.Error, err
	}

	switch con {
	case network.CanConnect:
		return net.CanConnect, nil
	case network.CannotConnect:
		return net.CannotConnect, nil
	case network.Connected:
		return net.Connected, nil
	case network.NotConnected:
		return net.NotConnected, nil
	default:
		return net.Unknown, nil
	}
}
