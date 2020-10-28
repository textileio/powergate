package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Wallet provides access to Powergate wallet admin APIs.
type Wallet struct {
	client proto.PowergateAdminServiceClient
}

// NewAddress creates a new address.
func (w *Wallet) NewAddress(ctx context.Context, addrType string) (*proto.NewAddressResponse, error) {
	req := &proto.NewAddressRequest{
		Type: addrType,
	}
	return w.client.NewAddress(ctx, req)
}

// Addresses lists all addresses associated with this Powergate.
func (w *Wallet) Addresses(ctx context.Context) (*proto.AddressesResponse, error) {
	return w.client.Addresses(ctx, &proto.AddressesRequest{})
}

// SendFil sends FIL from an address associated with this Powergate to any other address.
func (w *Wallet) SendFil(ctx context.Context, from, to string, amount int64) (*proto.SendFilResponse, error) {
	req := &proto.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return w.client.SendFil(ctx, req)
}
