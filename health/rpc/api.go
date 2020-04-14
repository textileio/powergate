package rpc

import (
	context "context"

	"github.com/textileio/powergate/health"
)

type API struct {
	UnimplementedAPIServer

	module *health.Module
}

// NewService creates a new Service
func NewService(m *health.Module) *API {
	return &API{module: m}
}

func (a *API) Check(ctx context.Context, req *CheckRequest) (*CheckReply, error) {
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
