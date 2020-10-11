package server

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// PowergateService implements the Powergate API.
type PowergateService struct {
}

// NewPowergateService creates a new AdminService.
func NewPowergateService() *PowergateService {
	return &PowergateService{}
}

// Balance returns the balance for an address.
func (p *PowergateService) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	return nil, nil
}
