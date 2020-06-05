package client

import (
	"context"

	h "github.com/textileio/powergate/health"
	"github.com/textileio/powergate/health/rpc"
)

// Health provides an API for checking node Health.
type Health struct {
	client rpc.RPCServiceClient
}

// Check returns the node health status and any related messages.
func (health *Health) Check(ctx context.Context) (h.Status, []string, error) {
	resp, err := health.client.Check(ctx, &rpc.CheckRequest{})
	if err != nil {
		return h.Error, nil, err
	}
	var status h.Status
	switch resp.Status {
	case rpc.Status_STATUS_OK:
		status = h.Ok
	case rpc.Status_STATUS_DEGRADED:
		status = h.Degraded
	case rpc.Status_STATUS_ERROR:
		status = h.Error
	default:
		status = h.Unspecified
	}
	return status, resp.Messages, nil
}
