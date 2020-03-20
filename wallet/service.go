package wallet

import (
	"context"

	pb "github.com/textileio/powergate/wallet/pb"
)

// Service implements the gprc service
type Service struct {
	pb.UnimplementedAPIServer

	Module *Module
}

// NewService creates a new Service
func NewService(m *Module) *Service {
	return &Service{Module: m}
}

// NewAddress creates a new wallet
func (s *Service) NewAddress(ctx context.Context, req *pb.NewAddressRequest) (*pb.NewAddressReply, error) {
	res, err := s.Module.NewAddress(ctx, req.GetType())
	if err != nil {
		return nil, err
	}
	return &pb.NewAddressReply{Address: res}, nil
}

// WalletBalance checks a wallet balance
func (s *Service) WalletBalance(ctx context.Context, req *pb.WalletBalanceRequest) (*pb.WalletBalanceReply, error) {
	res, err := s.Module.Balance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &pb.WalletBalanceReply{Balance: res}, nil
}
