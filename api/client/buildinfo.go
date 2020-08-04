package client

import (
	"context"

	"github.com/textileio/powergate/buildinfo/rpc"
)

// BuildInfo provides the BuildInfo API.
type BuildInfo struct {
	client rpc.RPCServiceClient
}

// BuildInfo returns build info about the server.
func (b *BuildInfo) BuildInfo(ctx context.Context) (*rpc.BuildInfoResponse, error) {
	return b.client.BuildInfo(ctx, &rpc.BuildInfoRequest{})
}
