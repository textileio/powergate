package rpc

import (
	context "context"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/net"
)

// API implements the rpc service
type API struct {
	UnimplementedAPIServer

	module net.Module
}

// NewService creates a new Service
func NewService(m net.Module) *API {
	return &API{module: m}
}

// ListenAddr calls module.ListenAddr
func (a *API) ListenAddr(ctx context.Context, req *ListenAddrRequest) (*ListenAddrReply, error) {
	addrInfo, err := a.module.ListenAddr(ctx)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(addrInfo.Addrs))
	for i, addr := range addrInfo.Addrs {
		addrs[i] = addr.String()
	}

	return &ListenAddrReply{
		AddrInfo: &PeerAddrInfo{
			ID:    addrInfo.ID.String(),
			Addrs: addrs,
		},
	}, nil
}

// Peers calls module.Peers
func (a *API) Peers(ctx context.Context, req *PeersRequest) (*PeersReply, error) {
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
				ID:    peer.AddrInfo.ID.String(),
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

	return &PeersReply{
		Peers: peerResults,
	}, nil
}

// FindPeer calls module.FindPeer
func (a *API) FindPeer(ctx context.Context, req *FindPeerRequest) (*FindPeerReply, error) {
	peerID, err := peer.Decode(req.PeerID)
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

	return &FindPeerReply{
		PeerInfo: &PeerInfo{
			AddrInfo: &PeerAddrInfo{
				ID:    peerInfo.AddrInfo.ID.String(),
				Addrs: addrs,
			},
			Location: &Location{
				Country:   peerInfo.Location.Country,
				Latitude:  peerInfo.Location.Latitude,
				Longitude: peerInfo.Location.Longitude,
			},
		},
	}, nil
}

// ConnectPeer calls module.ConnectPeer
func (a *API) ConnectPeer(ctx context.Context, req *ConnectPeerRequest) (*ConnectPeerReply, error) {
	addrs := make([]ma.Multiaddr, len(req.PeerInfo.Addrs))
	for i, addr := range req.PeerInfo.Addrs {
		ma, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		addrs[i] = ma
	}

	id, err := peer.Decode(req.PeerInfo.ID)
	if err != nil {
		return nil, err
	}

	addrInfo := peer.AddrInfo{
		ID:    id,
		Addrs: addrs,
	}

	if err := a.module.ConnectPeer(ctx, addrInfo); err != nil {
		return nil, err
	}

	return &ConnectPeerReply{}, nil
}

// DisconnectPeer calls module.DisconnectPeer
func (a *API) DisconnectPeer(ctx context.Context, req *DisconnectPeerRequest) (*DisconnectPeerReply, error) {
	peerID, err := peer.Decode(req.PeerID)
	if err != nil {
		return nil, err
	}

	if err := a.module.DisconnectPeer(ctx, peerID); err != nil {
		return nil, err
	}

	return &DisconnectPeerReply{}, nil
}

// Connectedness calls module.Connectedness
func (a *API) Connectedness(ctx context.Context, req *ConnectednessRequest) (*ConnectednessReply, error) {
	peerID, err := peer.Decode(req.PeerID)
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
		c = Connectedness_CanConnect
	case net.CannotConnect:
		c = Connectedness_CannotConnect
	case net.Connected:
		c = Connectedness_Connected
	case net.NotConnected:
		c = Connectedness_NotConnected
	case net.Unknown:
		c = Connectedness_Unknown
	case net.Error:
		c = Connectedness_Error
	default:
		c = Connectedness_Unknown
	}

	return &ConnectednessReply{
		Connectedness: c,
	}, nil
}
