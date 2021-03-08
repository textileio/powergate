package admin

import (
	"context"
	"math/big"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewAddress creates a new address.
func (a *Service) NewAddress(ctx context.Context, req *adminPb.NewAddressRequest) (*adminPb.NewAddressResponse, error) {
	addr, err := a.wm.NewAddress(ctx, req.AddressType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating address: %v", err)
	}
	return &adminPb.NewAddressResponse{
		Address: addr,
	}, nil
}

// Addresses lists all addresses associated with this Powergate.
func (a *Service) Addresses(ctx context.Context, req *adminPb.AddressesRequest) (*adminPb.AddressesResponse, error) {
	addrs, err := a.wm.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing addrs: %v", err)
	}
	return &adminPb.AddressesResponse{
		Addresses: addrs,
	}, nil
}

// AddressesDetailed lists all addresses with more detailed information such as balances and being verified-clients.
func (a *Service) AddressesDetailed(ctx context.Context, req *adminPb.AddressesDetailedRequest) (*adminPb.AddressesDetailedResponse, error) {
	addrs, err := a.wm.ListDetailed(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing addrs: %v", err)
	}

	res := &adminPb.AddressesDetailedResponse{
		Addresses: make([]*adminPb.AddressesDetailedResponse_AddressDetail, len(addrs)),
	}
	for i, addr := range addrs {
		daddr := &adminPb.AddressesDetailedResponse_AddressDetail{
			Address: addr.Address,
		}
		if addr.VCInfo != nil {
			daddr.VerifiedClientInfo = &adminPb.AddressesDetailedResponse_AddressDetail_VerifiedClientInfo{
				RemainingDatacapBytes: addr.VCInfo.RemainingDatacapBytes.String(),
			}
		}
		res.Addresses[i] = daddr
	}

	return res, nil
}

// SendFil sends FIL from an address associated with this Powergate to any other address.
func (a *Service) SendFil(ctx context.Context, req *adminPb.SendFilRequest) (*adminPb.SendFilResponse, error) {
	amt, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "parsing amount %v", req.Amount)
	}
	cid, err := a.wm.SendFil(ctx, req.From, req.To, amt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "sending fil: %v", err)
	}
	return &adminPb.SendFilResponse{
		Cid: cid.String(),
	}, nil
}
