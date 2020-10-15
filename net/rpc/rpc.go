package rpc

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/powergate/net"
)

// RPC implements the rpc service.
type RPC struct {
	UnimplementedRPCServiceServer

	module net.Module
}

// New creates a new rpc service.
func New(m net.Module) *RPC {
	return &RPC{module: m}
}

// ListenAddr calls module.ListenAddr.
func (a *RPC) ListenAddr(ctx context.Context, req *ListenAddrRequest) (*ListenAddrResponse, error) {
	addrInfo, err := a.module.ListenAddr(ctx)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(addrInfo.Addrs))
	for i, addr := range addrInfo.Addrs {
		addrs[i] = addr.String()
	}

	return &ListenAddrResponse{
		AddrInfo: &PeerAddrInfo{
			Id:    addrInfo.ID.String(),
			Addrs: addrs,
		},
	}, nil
}

// Peers calls module.Peers.
func (a *RPC) Peers(ctx context.Context, req *PeersRequest) (*PeersResponse, error) {
	peers, err := a.module.Peers(ctx)
	if err != nil {
		return nil, err
	}

	peerResults := make([]*PeerInfo, len(peers))
	for i, peer := range peers {
		addrs := make([]string, len(peer.AddrInfo.Addrs))
		for i, addr := range peer.AddrInfo.Addrs {
			addrs[i] = addr.String()
		}
		peerResults[i] = &PeerInfo{
			AddrInfo: &PeerAddrInfo{
				Id:    peer.AddrInfo.ID.String(),
				Addrs: addrs,
			},
		}
		if peer.Location != nil {
			peerResults[i].Location = &Location{
				Country:   peer.Location.Country,
				Latitude:  peer.Location.Latitude,
				Longitude: peer.Location.Longitude,
			}
		}
	}

	return &PeersResponse{
		Peers: peerResults,
	}, nil
}

// FindPeer calls module.FindPeer.
func (a *RPC) FindPeer(ctx context.Context, req *FindPeerRequest) (*FindPeerResponse, error) {
	peerID, err := peer.Decode(req.PeerId)
	if err != nil {
		return nil, err
	}
	peerInfo, err := a.module.FindPeer(ctx, peerID)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(peerInfo.AddrInfo.Addrs))
	for i, addr := range peerInfo.AddrInfo.Addrs {
		addrs[i] = addr.String()
	}

	var l *Location
	if peerInfo.Location != nil {
		l = &Location{
			Country:   peerInfo.Location.Country,
			Latitude:  peerInfo.Location.Latitude,
			Longitude: peerInfo.Location.Longitude,
		}
	}

	return &FindPeerResponse{
		PeerInfo: &PeerInfo{
			AddrInfo: &PeerAddrInfo{
				Id:    peerInfo.AddrInfo.ID.String(),
				Addrs: addrs,
			},
			Location: l,
		},
	}, nil
}

// Connectedness calls module.Connectedness.
func (a *RPC) Connectedness(ctx context.Context, req *ConnectednessRequest) (*ConnectednessResponse, error) {
	peerID, err := peer.Decode(req.PeerId)
	if err != nil {
		return nil, err
	}
	con, err := a.module.Connectedness(ctx, peerID)
	if err != nil {
		return nil, err
	}

	var c Connectedness
	switch con {
	case net.CanConnect:
		c = Connectedness_CONNECTEDNESS_CAN_CONNECT
	case net.CannotConnect:
		c = Connectedness_CONNECTEDNESS_CANNOT_CONNECT
	case net.Connected:
		c = Connectedness_CONNECTEDNESS_CONNECTED
	case net.NotConnected:
		c = Connectedness_CONNECTEDNESS_NOT_CONNECTED
	case net.Error:
		c = Connectedness_CONNECTEDNESS_ERROR
	default:
		c = Connectedness_CONNECTEDNESS_UNSPECIFIED
	}

	return &ConnectednessResponse{
		Connectedness: c,
	}, nil
}
