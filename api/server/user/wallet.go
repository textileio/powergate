package user

import (
	"context"
	"fmt"
	"math/big"

	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/ffs/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Balance returns the balance for an address.
func (s *Service) Balance(ctx context.Context, req *userPb.BalanceRequest) (*userPb.BalanceResponse, error) {
	bal, err := s.w.Balance(ctx, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting balance: %v", err)
	}
	return &userPb.BalanceResponse{Balance: bal.String()}, nil
}

// NewAddress calls ffs.NewAddr.
func (s *Service) NewAddress(ctx context.Context, req *userPb.NewAddressRequest) (*userPb.NewAddressResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	var opts []api.NewAddressOption
	if req.AddressType != "" {
		opts = append(opts, api.WithAddressType(req.AddressType))
	}
	if req.MakeDefault {
		opts = append(opts, api.WithMakeDefault(req.MakeDefault))
	}

	addr, err := i.NewAddr(ctx, req.Name, opts...)
	if err != nil {
		return nil, err
	}
	return &userPb.NewAddressResponse{Address: addr}, nil
}

// Addresses calls ffs.Addrs.
func (s *Service) Addresses(ctx context.Context, req *userPb.AddressesRequest) (*userPb.AddressesResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	addrs := i.Addrs()
	res := make([]*userPb.AddrInfo, len(addrs))
	for i, addr := range addrs {
		bal, err := s.w.Balance(ctx, addr.Addr)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "getting address balance: %v", err)
		}
		res[i] = &userPb.AddrInfo{
			Name:    addr.Name,
			Address: addr.Addr,
			Type:    addr.Type,
			Balance: bal.String(),
		}
	}
	return &userPb.AddressesResponse{Addresses: res}, nil
}

// SendFil sends fil from a managed address to any other address.
func (s *Service) SendFil(ctx context.Context, req *userPb.SendFilRequest) (*userPb.SendFilResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	amt, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "parsing amount %v", req.Amount)
	}
	if err := i.SendFil(ctx, req.From, req.To, amt); err != nil {
		return nil, err
	}
	return &userPb.SendFilResponse{}, nil
}

// SignMessage calls ffs.SignMessage.
func (s *Service) SignMessage(ctx context.Context, req *userPb.SignMessageRequest) (*userPb.SignMessageResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	signature, err := i.SignMessage(ctx, req.Address, req.Message)
	if err != nil {
		return nil, fmt.Errorf("signing message: %s", err)
	}

	return &userPb.SignMessageResponse{Signature: signature}, nil
}

// VerifyMessage calls ffs.VerifyMessage.
func (s *Service) VerifyMessage(ctx context.Context, req *userPb.VerifyMessageRequest) (*userPb.VerifyMessageResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := i.VerifyMessage(ctx, req.Address, req.Message, req.Signature)
	if err != nil {
		return nil, fmt.Errorf("verifying signature: %s", err)
	}

	return &userPb.VerifyMessageResponse{Ok: ok}, nil
}
