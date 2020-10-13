package powergate

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Balance returns the balance for an address.
func (p *Service) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	bal, err := p.w.Balance(ctx, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting balance: %v", err)
	}
	return &proto.BalanceResponse{Balance: bal}, nil
}
