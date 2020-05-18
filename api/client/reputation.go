package client

import (
	"context"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/reputation"
	"github.com/textileio/powergate/reputation/rpc"
)

// Reputation provides an API for viewing reputation data
type Reputation struct {
	client rpc.RPCClient
}

// AddSource adds a new external Source to be considered for reputation generation
func (r *Reputation) AddSource(ctx context.Context, id string, maddr ma.Multiaddr) error {
	req := &rpc.AddSourceRequest{
		Id:    id,
		Maddr: maddr.String(),
	}
	_, err := r.client.AddSource(ctx, req)
	return err
}

// GetTopMiners gets the top n miners with best score
func (r *Reputation) GetTopMiners(ctx context.Context, limit int) ([]reputation.MinerScore, error) {
	req := &rpc.GetTopMinersRequest{Limit: int32(limit)}
	reply, err := r.client.GetTopMiners(ctx, req)
	if err != nil {
		return []reputation.MinerScore{}, err
	}
	topMiners := make([]reputation.MinerScore, len(reply.GetTopMiners()))
	for i, val := range reply.GetTopMiners() {
		topMiners[i] = reputation.MinerScore{
			Addr:  val.GetAddr(),
			Score: int(val.GetScore()),
		}
	}
	return topMiners, nil
}
