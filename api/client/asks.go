package client

import (
	"context"

	"github.com/textileio/powergate/index/ask/rpc"
)

// Asks provides an API for viewing asks data.
type Asks struct {
	client rpc.RPCServiceClient
}

// Get returns the current index of available asks.
func (a *Asks) Get(ctx context.Context) (*rpc.GetResponse, error) {
	return a.client.Get(ctx, &rpc.GetRequest{})
}

// Query executes a query to retrieve active Asks.
func (a *Asks) Query(ctx context.Context, query *rpc.Query) (*rpc.QueryResponse, error) {
	return a.client.Query(ctx, &rpc.QueryRequest{Query: query})
}
