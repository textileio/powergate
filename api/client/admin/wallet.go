package admin

import (
	"context"
	"math/big"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Wallet provides access to Powergate wallet admin APIs.
type Wallet struct {
	client adminPb.AdminServiceClient
}

// NewAddress creates a new address.
func (w *Wallet) NewAddress(ctx context.Context, addrType string) (*adminPb.NewAddressResponse, error) {
	req := &adminPb.NewAddressRequest{
		AddressType: addrType,
	}
	return w.client.NewAddress(ctx, req)
}

// Addresses lists all addresses associated with this Powergate.
func (w *Wallet) Addresses(ctx context.Context) (*adminPb.AddressesResponse, error) {
	return w.client.Addresses(ctx, &adminPb.AddressesRequest{})
}

// SendFil sends FIL from an address associated with this Powergate to any other address.
func (w *Wallet) SendFil(ctx context.Context, from, to string, amount *big.Int) (*adminPb.SendFilResponse, error) {
	req := &adminPb.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount.String(),
	}
	return w.client.SendFil(ctx, req)
}
