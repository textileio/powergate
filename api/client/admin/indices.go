package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

type Indices struct {
	client adminPb.AdminServiceClient
}

func (i *Indices) GetMiners(ctx context.Context, opts ...GetMinersOption) (*adminPb.GetMinersResponse, error) {
	cfg := defaultCfgGetMiners
	for _, o := range opts {
		o(&cfg)
	}
	req := &adminPb.GetMinersRequest{
		WithPower: cfg.withPower,
	}

	return i.client.GetMiners(ctx, req)
}

func (i *Indices) GetMinerInfo(ctx context.Context, miners ...string) (*adminPb.GetMinerInfoResponse, error) {
	req := &adminPb.GetMinerInfoRequest{
		Miners: miners,
	}

	return i.client.GetMinerInfo(ctx, req)
}
