package wallet

import (
	"context"

	pb "github.com/textileio/fil-tools/wallet/pb"
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

// NewWallet creates a new wallet
func (s *Service) NewWallet(ctx context.Context, req *pb.NewWalletRequest) (*pb.NewWalletReply, error) {
	res, err := s.Module.NewWallet(ctx, req.GetTyp())
	if err != nil {
		return nil, err
	}
	return &pb.NewWalletReply{Address: res}, nil
}

// WalletBalance checks a wallet balance
func (s *Service) WalletBalance(ctx context.Context, req *pb.WalletBalanceRequest) (*pb.WalletBalanceReply, error) {
	res, err := s.Module.WalletBalance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &pb.WalletBalanceReply{Balance: res.Int64()}, nil
}
