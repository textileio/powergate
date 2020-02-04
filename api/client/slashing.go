package client

import (
	"context"

	"github.com/textileio/fil-tools/index/slashing"
	pb "github.com/textileio/fil-tools/index/slashing/pb"
)

// Slashing provides an API for viewing slashing data
type Slashing struct {
	client pb.APIClient
}

// Get returns the current index of miner slashes data
func (s *Slashing) Get(ctx context.Context) (*slashing.Index, error) {
	reply, err := s.client.Get(ctx, &pb.GetRequest{})
	if err != nil {
		return nil, err
	}

	miners := make(map[string]slashing.Slashes, len(reply.GetIndex().GetMiners()))
	for key, val := range reply.GetIndex().GetMiners() {
		miners[key] = slashing.Slashes{Epochs: val.GetEpochs()}
	}

	index := &slashing.Index{
		TipSetKey: reply.GetIndex().GetTipSetKey(),
		Miners:    miners,
	}

	return index, nil
}
