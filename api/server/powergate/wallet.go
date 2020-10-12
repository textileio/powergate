package powergate

import (
	"context"

	"github.com/filecoin-project/go-address"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Balance returns the balance for an address.
func (p *Service) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	client, cls, err := p.clientBuilder()
	defer cls()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating lotus client: %v", err)
	}
	a, err := address.NewFromString(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing address: %v", err)
	}
	bal, err := client.WalletBalance(ctx, a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting balance from lotus: %v", err)
	}
	return &proto.BalanceResponse{Balance: bal.Uint64()}, nil
}
