package net

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
)

type Module interface {
	ListenAddr(context.Context) (peer.AddrInfo, error)
	Peers(context.Context) ([]peer.AddrInfo, error)
	FindPeer(context.Context, peer.ID) (peer.AddrInfo, error)
	ConnectPeer(context.Context, peer.AddrInfo) error
	DisconnectPeer(context.Context, peer.ID) error
}
