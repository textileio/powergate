package client

import (
	"context"

	"github.com/textileio/powergate/index/faults"
	"github.com/textileio/powergate/index/faults/rpc"
)

// Faults provides an API for viewing faults data.
type Faults struct {
	client rpc.RPCServiceClient
}

// Get returns the current index of miner faults data.
func (s *Faults) Get(ctx context.Context) (*faults.IndexSnapshot, error) {
	reply, err := s.client.Get(ctx, &rpc.GetRequest{})
	if err != nil {
		return nil, err
	}

	miners := make(map[string]faults.Faults, len(reply.GetIndex().GetMiners()))
	for key, val := range reply.GetIndex().GetMiners() {
		miners[key] = faults.Faults{Epochs: val.GetEpochs()}
	}

	index := &faults.IndexSnapshot{
		TipSetKey: reply.GetIndex().GetTipsetkey(),
		Miners:    miners,
	}

	return index, nil
}
