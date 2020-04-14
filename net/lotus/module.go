package lotus

import (
	"context"

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
func New(api *apistruct.FullNodeStruct) (*Module, error) {
	m := &Module{
		api: api,
	}
	return m, nil
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
