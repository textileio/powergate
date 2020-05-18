package client

import (
	"context"

	"github.com/textileio/powergate/index/slashing"
	"github.com/textileio/powergate/index/slashing/rpc"
)

// Slashing provides an API for viewing slashing data
type Slashing struct {
	client rpc.APIClient
}

// Get returns the current index of miner slashes data
func (s *Slashing) Get(ctx context.Context) (*slashing.IndexSnapshot, error) {
	reply, err := s.client.Get(ctx, &rpc.GetRequest{})
	if err != nil {
		return nil, err
	}

	miners := make(map[string]slashing.Slashes, len(reply.GetIndex().GetMiners()))
	for key, val := range reply.GetIndex().GetMiners() {
		miners[key] = slashing.Slashes{Epochs: val.GetEpochs()}
	}

	index := &slashing.IndexSnapshot{
		TipSetKey: reply.GetIndex().GetTipSetKey(),
		Miners:    miners,
	}

	return index, nil
}
