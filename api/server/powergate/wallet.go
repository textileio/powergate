package powergate

import (
	"context"
	"fmt"
	"math/big"

	"github.com/textileio/powergate/ffs/api"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Balance returns the balance for an address.
func (s *Service) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	bal, err := s.w.Balance(ctx, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting balance: %v", err)
	}
	return &proto.BalanceResponse{Balance: bal}, nil
}

// NewAddr calls ffs.NewAddr.
func (s *Service) NewAddr(ctx context.Context, req *proto.NewAddrRequest) (*proto.NewAddrResponse, error) {
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
	return &proto.NewAddrResponse{Addr: addr}, nil
}

// Addrs calls ffs.Addrs.
func (s *Service) Addrs(ctx context.Context, req *proto.AddrsRequest) (*proto.AddrsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}
	addrs := i.Addrs()
	res := make([]*proto.AddrInfo, len(addrs))
	for i, addr := range addrs {
		bal, err := s.w.Balance(ctx, addr.Addr)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "getting address balance: %v", err)
		}
		res[i] = &proto.AddrInfo{
			Name:    addr.Name,
			Addr:    addr.Addr,
			Type:    addr.Type,
			Balance: bal,
		}
	}
	return &proto.AddrsResponse{Addrs: res}, nil
}

// SendFil sends fil from a managed address to any other address.
func (s *Service) SendFil(ctx context.Context, req *proto.SendFilRequest) (*proto.SendFilResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := i.SendFil(ctx, req.From, req.To, big.NewInt(req.Amount)); err != nil {
		return nil, err
	}
	return &proto.SendFilResponse{}, nil
}

// SignMessage calls ffs.SignMessage.
func (s *Service) SignMessage(ctx context.Context, req *proto.SignMessageRequest) (*proto.SignMessageResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	signature, err := i.SignMessage(ctx, req.Addr, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("signing message: %s", err)
	}

	return &proto.SignMessageResponse{Signature: signature}, nil
}

// VerifyMessage calls ffs.VerifyMessage.
func (s *Service) VerifyMessage(ctx context.Context, req *proto.VerifyMessageRequest) (*proto.VerifyMessageResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := i.VerifyMessage(ctx, req.Addr, req.Msg, req.Signature)
	if err != nil {
		return nil, fmt.Errorf("verifying signature: %s", err)
	}

	return &proto.VerifyMessageResponse{Ok: ok}, nil
}
