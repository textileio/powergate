package health

import (
	"context"
	"fmt"

	"github.com/textileio/powergate/net"
)

// Module exposes the health api.
type Module struct {
	net net.Module
}

// Status represents the node's health status
type Status int

const (
	// Ok specifies the node is healthy
	Ok Status = iota
	// Degraded specifies there are problems with the node health
	Degraded
	// Error specifies there was an error when determining node health
	Error
)

// New creates a new net module
func New(net net.Module) *Module {
	m := &Module{
		net: net,
	}
	return m
}

// Check reutuns the current health status and any messages related to the status
func (m *Module) Check(ctx context.Context) (status Status, messages []string, err error) {
	peers, err := m.net.Peers(ctx)
	if err != nil {
		return Error, nil, err
	}
	for _, peer := range peers {
		con, err := m.net.Connectedness(ctx, peer.AddrInfo.ID)
		if err != nil {
			messages = append(messages, fmt.Sprintf("error checking connectedness for peer %v: %v", peer.AddrInfo.ID.String(), err))
			continue
		}
		if con == net.CannotConnect || con == net.Unknown || con == net.Error {
			messages = append(messages, fmt.Sprintf("degraded connectedness %v for peer %v", con, peer.AddrInfo.ID.String()))
		}
	}
	status = Ok
	if len(messages) > 0 {
		status = Degraded
	}
	return status, messages, nil
}
