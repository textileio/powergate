package client

import (
	"context"
	"math/big"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// Wallet provides an API for managing filecoin wallets.
type Wallet struct {
	client userPb.UserServiceClient
}

// NewAddressOption is a function that changes a NewAddressConfig.
type NewAddressOption func(r *userPb.NewAddressRequest)

// WithMakeDefault specifies if the new address should become the default.
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(r *userPb.NewAddressRequest) {
		r.MakeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create.
func WithAddressType(addressType string) NewAddressOption {
	return func(r *userPb.NewAddressRequest) {
		r.AddressType = addressType
	}
}

// Balance gets a filecoin wallet's balance.
func (w *Wallet) Balance(ctx context.Context, address string) (*userPb.BalanceResponse, error) {
	return w.client.Balance(ctx, &userPb.BalanceRequest{Address: address})
}

// NewAddress creates a new wallet address managed by the user.
func (w *Wallet) NewAddress(ctx context.Context, name string, options ...NewAddressOption) (*userPb.NewAddressResponse, error) {
	r := &userPb.NewAddressRequest{Name: name}
	for _, opt := range options {
		opt(r)
	}
	return w.client.NewAddress(ctx, r)
}

// Addresses returns a list of addresses managed by the user.
func (w *Wallet) Addresses(ctx context.Context) (*userPb.AddressesResponse, error) {
	return w.client.Addresses(ctx, &userPb.AddressesRequest{})
}

// SendFil sends fil from a managed address to any another address, returns immediately but funds are sent asynchronously.
func (w *Wallet) SendFil(ctx context.Context, from string, to string, amount *big.Int) (*userPb.SendFilResponse, error) {
	req := &userPb.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount.String(),
	}
	return w.client.SendFil(ctx, req)
}

// SignMessage signs a message with a user wallet address.
func (w *Wallet) SignMessage(ctx context.Context, address string, message []byte) (*userPb.SignMessageResponse, error) {
	r := &userPb.SignMessageRequest{Address: address, Message: message}
	return w.client.SignMessage(ctx, r)
}

// VerifyMessage verifies a message signature from a wallet address.
func (w *Wallet) VerifyMessage(ctx context.Context, address string, message, signature []byte) (*userPb.VerifyMessageResponse, error) {
	r := &userPb.VerifyMessageRequest{Address: address, Message: message, Signature: signature}
	return w.client.VerifyMessage(ctx, r)
}
