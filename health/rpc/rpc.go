package rpc

import (
	context "context"

	"github.com/textileio/powergate/health"
)

// RPC implements the rpc service.
type RPC struct {
	UnimplementedRPCServiceServer

	module *health.Module
}

// New creates a new rpc service.
func New(m *health.Module) *RPC {
	return &RPC{module: m}
}

// Check calls module.Check.
func (a *RPC) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	status, messages, err := a.module.Check(ctx)
	if err != nil {
		return nil, err
	}

	var s Status
	switch status {
	case health.Ok:
		s = Status_STATUS_OK
	case health.Degraded:
		s = Status_STATUS_DEGRADED
	case health.Error:
		s = Status_STATUS_ERROR
	default:
		s = Status_STATUS_UNSPECIFIED
	}

	return &CheckResponse{
		Status:   s,
		Messages: messages,
	}, nil
}
