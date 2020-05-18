package client

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/iplocation"
	n "github.com/textileio/powergate/net"
	"github.com/textileio/powergate/net/rpc"
)

// Net provides the Net API
type Net struct {
	client rpc.APIClient
}

// ListenAddr returns listener address info for the local node
func (net *Net) ListenAddr(ctx context.Context) (peer.AddrInfo, error) {
	resp, err := net.client.ListenAddr(ctx, &rpc.ListenAddrRequest{})
	if err != nil {
		return peer.AddrInfo{}, err
	}
	addrs := make([]ma.Multiaddr, len(resp.AddrInfo.Addrs))
	for i, addr := range resp.AddrInfo.Addrs {
		ma, err := ma.NewMultiaddr(addr)
		if err != nil {
			return peer.AddrInfo{}, err
		}
		addrs[i] = ma
	}
	id, err := peer.Decode(resp.AddrInfo.ID)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return peer.AddrInfo{
		ID:    id,
		Addrs: addrs,
	}, nil
}

// Peers returns a list of peers
func (net *Net) Peers(ctx context.Context) ([]n.PeerInfo, error) {
	resp, err := net.client.Peers(ctx, &rpc.PeersRequest{})
	if err != nil {
		return nil, err
	}
	peerInfos := make([]n.PeerInfo, len(resp.Peers))
	for i, p := range resp.Peers {
		peerInfo, err := fromProtoPeerInfo(p)
		if err != nil {
			return nil, err
		}
		peerInfos[i] = peerInfo
	}
	return peerInfos, nil
}

// FindPeer finds a peer by peer id
func (net *Net) FindPeer(ctx context.Context, peerID peer.ID) (n.PeerInfo, error) {
	resp, err := net.client.FindPeer(ctx, &rpc.FindPeerRequest{PeerID: peerID.String()})
	if err != nil {
		return n.PeerInfo{}, err
	}
	return fromProtoPeerInfo(resp.PeerInfo)
}

// ConnectPeer connects to a peer
func (net *Net) ConnectPeer(ctx context.Context, addrInfo peer.AddrInfo) error {
	addrs := make([]string, len(addrInfo.Addrs))
	for i, addr := range addrInfo.Addrs {
		addrs[i] = addr.String()
	}
	info := &rpc.PeerAddrInfo{
		ID:    addrInfo.ID.String(),
		Addrs: addrs,
	}
	_, err := net.client.ConnectPeer(ctx, &rpc.ConnectPeerRequest{PeerInfo: info})
	return err
}

// DisconnectPeer disconnects from a peer
func (net *Net) DisconnectPeer(ctx context.Context, peerID peer.ID) error {
	_, err := net.client.DisconnectPeer(ctx, &rpc.DisconnectPeerRequest{PeerID: peerID.String()})
	return err
}

// Connectedness returns the connection status to a peer
func (net *Net) Connectedness(ctx context.Context, peerID peer.ID) (n.Connectedness, error) {
	resp, err := net.client.Connectedness(ctx, &rpc.ConnectednessRequest{PeerID: peerID.String()})
	if err != nil {
		return n.Error, err
	}
	var con n.Connectedness
	switch resp.Connectedness {
	case rpc.Connectedness_CanConnect:
		con = n.CanConnect
	case rpc.Connectedness_CannotConnect:
		con = n.CannotConnect
	case rpc.Connectedness_Connected:
		con = n.Connected
	case rpc.Connectedness_NotConnected:
		con = n.NotConnected
	case rpc.Connectedness_Error:
		con = n.Error
	case rpc.Connectedness_Unknown:
		con = n.Unknown
	default:
		con = n.Unknown
	}
	return con, nil
}

func fromProtoPeerInfo(proto *rpc.PeerInfo) (n.PeerInfo, error) {
	addrs := make([]ma.Multiaddr, len(proto.AddrInfo.Addrs))
	for i, addr := range proto.AddrInfo.Addrs {
		ma, err := ma.NewMultiaddr(addr)
		if err != nil {
			return n.PeerInfo{}, err
		}
		addrs[i] = ma
	}
	id, err := peer.Decode(proto.AddrInfo.ID)
	if err != nil {
		return n.PeerInfo{}, err
	}
	peerInfo := n.PeerInfo{
		AddrInfo: peer.AddrInfo{
			ID:    id,
			Addrs: addrs,
		},
	}
	if proto.Location != nil {
		peerInfo.Location = &iplocation.Location{
			Country:   proto.Location.Country,
			Latitude:  proto.Location.Latitude,
			Longitude: proto.Location.Longitude,
		}
	}

	return peerInfo, nil
}
