package net

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/powergate/iplocation"
)

// Connectedness signals the capacity for a connection with a given node
type Connectedness int

const (
	// NotConnected means no connection to peer, and no extra information (default)
	NotConnected Connectedness = iota

	// Connected means has an open, live connection to peer
	Connected

	// CanConnect means recently connected to peer, terminated gracefully
	CanConnect

	// CannotConnect means recently attempted connecting but failed to connect.
	CannotConnect

	// Unknown means unable to determine connectedness
	Unknown

	// Error means there was an error determining connectedness
	Error
)

type PeerInfo struct {
	AddrInfo peer.AddrInfo
	Location iplocation.Location
}

type Module interface {
	ListenAddr(context.Context) (peer.AddrInfo, error)
	Peers(context.Context) ([]PeerInfo, error)
	FindPeer(context.Context, peer.ID) (PeerInfo, error)
	ConnectPeer(context.Context, peer.AddrInfo) error
	DisconnectPeer(context.Context, peer.ID) error
	Connectedness(context.Context, peer.ID) (Connectedness, error)
}
