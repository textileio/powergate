package net

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
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

type Module interface {
	ListenAddr(context.Context) (peer.AddrInfo, error)
	Peers(context.Context) ([]peer.AddrInfo, error)
	FindPeer(context.Context, peer.ID) (peer.AddrInfo, error)
	ConnectPeer(context.Context, peer.AddrInfo) error
	DisconnectPeer(context.Context, peer.ID) error
	Connectedness(context.Context, peer.ID) (Connectedness, error)
}
