package lotus

import (
	"context"
	"fmt"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/powergate/iplocation"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/net"
)

var (
	_ net.Module = (*Module)(nil)

	log = logging.Logger("net")
)

// Module exposes the filecoin wallet api.
type Module struct {
	clientBuilder lotus.ClientBuilder
	lr            iplocation.LocationResolver
}

// New creates a new net module.
func New(clientBuilder lotus.ClientBuilder, lr iplocation.LocationResolver) *Module {
	m := &Module{
		clientBuilder: clientBuilder,
		lr:            lr,
	}
	return m
}

// ListenAddr implements ListenAddr.
func (m *Module) ListenAddr(ctx context.Context) (peer.AddrInfo, error) {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	return client.NetAddrsListen(ctx)
}

// ConnectPeer implements ConnectPeer.
func (m *Module) ConnectPeer(ctx context.Context, addrInfo peer.AddrInfo) error {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	return client.NetConnect(ctx, addrInfo)
}

// DisconnectPeer implements DisconnectPeer.
func (m *Module) DisconnectPeer(ctx context.Context, peerID peer.ID) error {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	return client.NetDisconnect(ctx, peerID)
}

// FindPeer implements FindPeer.
func (m *Module) FindPeer(ctx context.Context, peerID peer.ID) (net.PeerInfo, error) {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return net.PeerInfo{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	addrInfo, err := client.NetFindPeer(ctx, peerID)
	if err != nil {
		return net.PeerInfo{}, err
	}
	loc, err := m.lr.Resolve(addrInfo.Addrs)
	if err == iplocation.ErrCantResolve {
		return net.PeerInfo{
			AddrInfo: addrInfo,
		}, nil
	}
	if err != nil {
		return net.PeerInfo{}, err
	}
	return net.PeerInfo{
		AddrInfo: addrInfo,
		Location: &loc,
	}, nil
}

// Peers implements Peers.
func (m *Module) Peers(ctx context.Context) ([]net.PeerInfo, error) {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return nil, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	addrInfos, err := client.NetPeers(ctx)
	if err != nil {
		return nil, err
	}
	peerInfos := make([]net.PeerInfo, len(addrInfos))
	for i, addrInfo := range addrInfos {
		loc, err := m.lr.Resolve(addrInfo.Addrs)
		if err == iplocation.ErrCantResolve {
			peerInfos[i] = net.PeerInfo{
				AddrInfo: addrInfo,
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		peerInfos[i] = net.PeerInfo{
			AddrInfo: addrInfo,
			Location: &loc,
		}
	}
	return peerInfos, nil
}

// Connectedness implements Connectedness.
func (m *Module) Connectedness(ctx context.Context, peerID peer.ID) (net.Connectedness, error) {
	client, cls, err := m.clientBuilder(context.Background())
	if err != nil {
		return net.Error, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	con, err := client.NetConnectedness(ctx, peerID)
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
		log.Warnf("unspecified peer connectedness &v", con)
		return net.Unspecified, nil
	}
}
