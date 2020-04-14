package rpc

import (
	context "context"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/net"
)

type API struct {
	UnimplementedAPIServer

	module net.Module
}

// NewService creates a new Service
func NewService(m net.Module) *API {
	return &API{module: m}
}

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
		AddrInfo: &AddrInfo{
			ID:    addrInfo.ID.String(),
			Addrs: addrs,
		},
	}, nil
}

func (a *API) Peers(ctx context.Context, req *PeersRequest) (*PeersReply, error) {
	peers, err := a.module.Peers(ctx)
	if err != nil {
		return nil, err
	}

	peerResults := make([]*AddrInfo, len(peers))
	for i, peer := range peers {
		addrs := make([]string, len(peer.Addrs))
		for i, addr := range peer.Addrs {
			addrs[i] = addr.String()
		}
		peerResults[i] = &AddrInfo{
			ID:    peer.ID.String(),
			Addrs: addrs,
		}
	}

	return &PeersReply{
		Peers: peerResults,
	}, nil
}

func (a *API) FindPeer(ctx context.Context, req *FindPeerRequest) (*FindPeerReply, error) {
	peerId := peer.ID(req.PeerID)
	addrInfo, err := a.module.FindPeer(ctx, peerId)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, len(addrInfo.Addrs))
	for i, addr := range addrInfo.Addrs {
		addrs[i] = addr.String()
	}

	return &FindPeerReply{
		PeerInfo: &AddrInfo{
			ID:    addrInfo.ID.String(),
			Addrs: addrs,
		},
	}, nil
}

func (a *API) ConnectPeer(ctx context.Context, req *ConnectPeerRequest) (*ConnectPeerReply, error) {
	addrs := make([]ma.Multiaddr, len(req.PeerInfo.Addrs))
	for i, addr := range req.PeerInfo.Addrs {
		ma, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		addrs[i] = ma
	}

	addrInfo := peer.AddrInfo{
		ID:    peer.ID(req.PeerInfo.ID),
		Addrs: addrs,
	}

	if err := a.module.ConnectPeer(ctx, addrInfo); err != nil {
		return nil, err
	}

	return &ConnectPeerReply{}, nil
}

func (a *API) DisconnectPeer(ctx context.Context, req *DisconnectPeerRequest) (*DisconnectPeerReply, error) {
	peerID := peer.ID(req.PeerID)

	if err := a.module.DisconnectPeer(ctx, peerID); err != nil {
		return nil, err
	}

	return &DisconnectPeerReply{}, nil
}

func (a *API) Connectedness(ctx context.Context, req *ConnectednessRequest) (*ConnectednessReply, error) {
	peerID := peer.ID(req.PeerID)
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
