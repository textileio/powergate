package rpc

import (
	context "context"

	"github.com/textileio/powergate/health"
)

// RPC implements the rpc service
type RPC struct {
	UnimplementedRPCServer

	module *health.Module
}

// New creates a new rpc service
func New(m *health.Module) *RPC {
	return &RPC{module: m}
}

// Check calls module.Check
func (a *RPC) Check(ctx context.Context, req *CheckRequest) (*CheckReply, error) {
	status, messages, err := a.module.Check(ctx)
	if err != nil {
		return nil, err
	}

	var s Status
	switch status {
	case health.Ok:
		s = Status_Ok
	case health.Degraded:
		s = Status_Degraded
	case health.Error:
		s = Status_Error
	}

	return &CheckReply{
		Status:   s,
		Messages: messages,
	}, nil
}
