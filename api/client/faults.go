package client

import (
	"context"

	"github.com/textileio/powergate/index/faults/rpc"
)

// Faults provides an API for viewing faults data.
type Faults struct {
	client rpc.RPCServiceClient
}

// Get returns the current index of miner faults data.
func (s *Faults) Get(ctx context.Context) (*rpc.GetResponse, error) {
	return s.client.Get(ctx, &rpc.GetRequest{})
}
