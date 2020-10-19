package client

import (
	"context"

	"github.com/textileio/powergate/reputation/rpc"
)

// Reputation provides an API for viewing reputation data.
type Reputation struct {
	client rpc.RPCServiceClient
}

// AddSource adds a new external Source to be considered for reputation generation.
func (r *Reputation) AddSource(ctx context.Context, id, maddr string) (*rpc.AddSourceResponse, error) {
	req := &rpc.AddSourceRequest{
		Id:    id,
		Maddr: maddr,
	}
	return r.client.AddSource(ctx, req)
}

// GetTopMiners gets the top n miners with best score.
func (r *Reputation) GetTopMiners(ctx context.Context, limit int32) (*rpc.GetTopMinersResponse, error) {
	req := &rpc.GetTopMinersRequest{Limit: limit}
	return r.client.GetTopMiners(ctx, req)
}
