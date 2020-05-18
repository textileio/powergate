package rpc

import (
	"context"

	"github.com/textileio/powergate/wallet"
)

// Service implements the gprc service
type Service struct {
	UnimplementedAPIServer

	Module *wallet.Module
}

// NewService creates a new Service
func NewService(m *wallet.Module) *Service {
	return &Service{Module: m}
}

// NewAddress creates a new wallet
func (s *Service) NewAddress(ctx context.Context, req *NewAddressRequest) (*NewAddressReply, error) {
	res, err := s.Module.NewAddress(ctx, req.GetType())
	if err != nil {
		return nil, err
	}
	return &NewAddressReply{Address: res}, nil
}

// WalletBalance checks a wallet balance
func (s *Service) WalletBalance(ctx context.Context, req *WalletBalanceRequest) (*WalletBalanceReply, error) {
	res, err := s.Module.Balance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &WalletBalanceReply{Balance: res}, nil
}
