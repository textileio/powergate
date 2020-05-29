package rpc

import (
	"context"

	"github.com/textileio/powergate/wallet"
)

// RPC implements the gprc service
type RPC struct {
	UnimplementedRPCServiceServer

	Module *wallet.Module
}

// New creates a new rpc service
func New(m *wallet.Module) *RPC {
	return &RPC{Module: m}
}

// NewAddress creates a new wallet
func (s *RPC) NewAddress(ctx context.Context, req *NewAddressRequest) (*NewAddressResponse, error) {
	res, err := s.Module.NewAddress(ctx, req.GetType())
	if err != nil {
		return nil, err
	}
	return &NewAddressResponse{Address: res}, nil
}

// WalletBalance checks a wallet balance
func (s *RPC) WalletBalance(ctx context.Context, req *WalletBalanceRequest) (*WalletBalanceResponse, error) {
	res, err := s.Module.Balance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &WalletBalanceResponse{Balance: res}, nil
}
