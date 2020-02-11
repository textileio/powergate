package client

import (
	"context"

	pb "github.com/textileio/fil-tools/index/slashing/pb"
	"github.com/textileio/fil-tools/index/slashing/types"
)

// Slashing provides an API for viewing slashing data
type Slashing struct {
	client pb.APIClient
}

// Get returns the current index of miner slashes data
func (s *Slashing) Get(ctx context.Context) (*types.Index, error) {
	reply, err := s.client.Get(ctx, &pb.GetRequest{})
	if err != nil {
		return nil, err
	}

	miners := make(map[string]types.Slashes, len(reply.GetIndex().GetMiners()))
	for key, val := range reply.GetIndex().GetMiners() {
		miners[key] = types.Slashes{Epochs: val.GetEpochs()}
	}

	index := &types.Index{
		TipSetKey: reply.GetIndex().GetTipSetKey(),
		Miners:    miners,
	}

	return index, nil
}
