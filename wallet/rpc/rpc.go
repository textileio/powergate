package rpc

import (
	"context"

	"github.com/textileio/powergate/wallet"
)

// RPC implements the gprc service.
type RPC struct {
	UnimplementedRPCServiceServer

	Module wallet.Module
}

// New creates a new rpc service.
func New(m wallet.Module) *RPC {
	return &RPC{Module: m}
}

// NewAddress creates a new wallet.
func (s *RPC) NewAddress(ctx context.Context, req *NewAddressRequest) (*NewAddressResponse, error) {
	res, err := s.Module.NewAddress(ctx, req.GetType())
	if err != nil {
		return nil, err
	}
	return &NewAddressResponse{Address: res}, nil
}

// List returns all wallet addresses.
func (s *RPC) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	res, err := s.Module.List(ctx)
	if err != nil {
		return nil, err
	}
	return &ListResponse{Addresses: res}, nil
}

// Balance checks a wallet balance.
func (s *RPC) Balance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error) {
	res, err := s.Module.Balance(ctx, req.GetAddress())
	if err != nil {
		return nil, err
	}
	return &BalanceResponse{Balance: res}, nil
}
