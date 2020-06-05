package net

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/textileio/powergate/iplocation"
)

// Connectedness signals the capacity for a connection with a given node.
type Connectedness int

const (
	// Unspecified means unable to determine connectedness.
	Unspecified Connectedness = iota

	// NotConnected means no connection to peer, and no extra information (default).
	NotConnected

	// Connected means has an open, live connection to peer.
	Connected

	// CanConnect means recently connected to peer, terminated gracefully.
	CanConnect

	// CannotConnect means recently attempted connecting but failed to connect.
	CannotConnect

	// Error means there was an error determining connectedness.
	Error
)

func (s Connectedness) String() string {
	names := [...]string{
		"Not Connected",
		"Connected",
		"Can Connect",
		"Cannot Connect",
		"Unknown",
		"Error",
	}
	if s < NotConnected || s > Error {
		return "Unknown"
	}
	return names[s]
}

// PeerInfo provides address info and location info about a peer.
type PeerInfo struct {
	AddrInfo peer.AddrInfo
	Location *iplocation.Location
}

//Module defines the net API.
type Module interface {
	// ListenAddr returns listener address info for the local node.
	ListenAddr(context.Context) (peer.AddrInfo, error)
	// Peers returns a list of peers.
	Peers(context.Context) ([]PeerInfo, error)
	// FindPeer finds a peer by peer id
	FindPeer(context.Context, peer.ID) (PeerInfo, error)
	// ConnectPeer connects to a peer.
	ConnectPeer(context.Context, peer.AddrInfo) error
	// DisconnectPeer disconnects from a peer.
	DisconnectPeer(context.Context, peer.ID) error
	// Connectedness returns the connection status to a peer.
	Connectedness(context.Context, peer.ID) (Connectedness, error)
}
