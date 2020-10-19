package client

import (
	"context"

	"github.com/textileio/powergate/health/rpc"
)

// Health provides an API for checking node Health.
type Health struct {
	client rpc.RPCServiceClient
}

// Check returns the node health status and any related messages.
func (health *Health) Check(ctx context.Context) (*rpc.CheckResponse, error) {
	return health.client.Check(ctx, &rpc.CheckRequest{})
}
