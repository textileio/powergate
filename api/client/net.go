package client

import (
	"context"

	"github.com/textileio/powergate/net/rpc"
)

// Net provides the Net API.
type Net struct {
	client rpc.RPCServiceClient
}

// ListenAddr returns listener address info for the local node.
func (net *Net) ListenAddr(ctx context.Context) (*rpc.ListenAddrResponse, error) {
	return net.client.ListenAddr(ctx, &rpc.ListenAddrRequest{})
}

// Peers returns a list of peers.
func (net *Net) Peers(ctx context.Context) (*rpc.PeersResponse, error) {
	return net.client.Peers(ctx, &rpc.PeersRequest{})
}

// FindPeer finds a peer by peer id.
func (net *Net) FindPeer(ctx context.Context, peerID string) (*rpc.FindPeerResponse, error) {
	return net.client.FindPeer(ctx, &rpc.FindPeerRequest{PeerId: peerID})
}

// Connectedness returns the connection status to a peer.
func (net *Net) Connectedness(ctx context.Context, peerID string) (*rpc.ConnectednessResponse, error) {
	return net.client.Connectedness(ctx, &rpc.ConnectednessRequest{PeerId: peerID})
}
