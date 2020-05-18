package rpc

import (
	"context"

	"github.com/textileio/powergate/wallet"
)

// RPC implements the gprc service
type RPC struct {
	UnimplementedRPCServer

	Module *wallet.Module
}

// New creates a new rpc service
func New(m *wallet.Module) *RPC {
	return &RPC{Module: m}
}

// NewAddress creates a new wallet
func (s *RPC) NewAddress(ctx context.Context, req *NewAddressRequest) (*NewAddressReply, error) {
	res, err := s.Module.NewAddress(ctx, req.GetType())
	if err != nil {
		return nil, err
	}
	return &NewAddressReply{Address: res}, nil
}

// WalletBalance checks a wallet balance
func (s *RPC) WalletBalance(ctx context.Context, req *WalletBalanceRequest) (*WalletBalanceReply, error) {
	res, err := s.Module.Balance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &WalletBalanceReply{Balance: res}, nil
}
