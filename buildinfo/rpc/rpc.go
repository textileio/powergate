package rpc

import (
	"context"

	"github.com/textileio/powergate/buildinfo"
)

// RPC implements the rpc service.
type RPC struct {
}

// New creates a new rpc service.
func New() *RPC {
	return &RPC{}
}

// BuildInfo returns information about the powergate build.
func (i *RPC) BuildInfo(ctx context.Context, req *BuildInfoRequest) (*BuildInfoResponse, error) {
	return &BuildInfoResponse{
		BuildDate:  buildinfo.BuildDate,
		GitBranch:  buildinfo.GitBranch,
		GitCommit:  buildinfo.GitCommit,
		GitState:   buildinfo.GitState,
		GitSummary: buildinfo.GitSummary,
		Version:    buildinfo.Version,
	}, nil
}
