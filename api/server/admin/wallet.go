package admin

import (
	"context"
	"math/big"

	proto "github.com/textileio/powergate/proto/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewAddress creates a new address.
func (a *Service) NewAddress(ctx context.Context, req *proto.NewAddressRequest) (*proto.NewAddressResponse, error) {
	addr, err := a.wm.NewAddress(ctx, req.AddressType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating address: %v", err)
	}
	return &proto.NewAddressResponse{
		Address: addr,
	}, nil
}

// Addresses lists all addresses associated with this Powergate.
func (a *Service) Addresses(ctx context.Context, req *proto.AddressesRequest) (*proto.AddressesResponse, error) {
	addrs, err := a.wm.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing addrs: %v", err)
	}
	return &proto.AddressesResponse{
		Addresses: addrs,
	}, nil
}

// SendFil sends FIL from an address associated with this Powergate to any other address.
func (a *Service) SendFil(ctx context.Context, req *proto.SendFilRequest) (*proto.SendFilResponse, error) {
	amt, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "parsing amount %v", req.Amount)
	}
	err := a.wm.SendFil(ctx, req.From, req.To, amt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "sending fil: %v", err)
	}
	return &proto.SendFilResponse{}, nil
}
