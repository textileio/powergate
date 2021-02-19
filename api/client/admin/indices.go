package admin

import (
	"context"
	"fmt"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Indices provides APIs to fetch indices data.
type Indices struct {
	client adminPb.AdminServiceClient
}

// GetMiners return a list of miner addresses that satisfies the provided filters.
func (i *Indices) GetMiners(ctx context.Context, opts ...GetMinersOption) (*adminPb.GetMinersResponse, error) {
	cfg := defaultCfgGetMiners
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return nil, fmt.Errorf("applying configuration option: %s", err)
		}
	}
	req := &adminPb.GetMinersRequest{
		WithPower: cfg.withPower,
	}

	return i.client.GetMiners(ctx, req)
}

// GetMinerInfo returns all known indices data for the provider miner addresses.
func (i *Indices) GetMinerInfo(ctx context.Context, miners ...string) (*adminPb.GetMinerInfoResponse, error) {
	req := &adminPb.GetMinerInfoRequest{
		Miners: miners,
	}

	return i.client.GetMinerInfo(ctx, req)
}
