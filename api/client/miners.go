package client

import (
	"context"

	"github.com/textileio/powergate/index/miner/rpc"
)

// Miners provides an API for viewing miner data.
type Miners struct {
	client rpc.RPCServiceClient
}

// Get returns the current index of available asks.
func (a *Miners) Get(ctx context.Context) (*rpc.GetResponse, error) {
	return a.client.Get(ctx, &rpc.GetRequest{})
}
