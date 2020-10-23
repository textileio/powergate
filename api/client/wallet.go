package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// Wallet provides an API for managing filecoin wallets.
type Wallet struct {
	client proto.PowergateServiceClient
}

// NewAddressOption is a function that changes a NewAddressConfig.
type NewAddressOption func(r *proto.NewAddrRequest)

// WithMakeDefault specifies if the new address should become the default.
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(r *proto.NewAddrRequest) {
		r.MakeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create.
func WithAddressType(addressType string) NewAddressOption {
	return func(r *proto.NewAddrRequest) {
		r.AddressType = addressType
	}
}

// // List returns all wallet addresses.
// func (w *Wallet) List(ctx context.Context) (*walletRpc.ListResponse, error) {
// 	return w.walletClient.List(ctx, &walletRpc.ListRequest{})
// }

// Balance gets a filecoin wallet's balance.
func (w *Wallet) Balance(ctx context.Context, address string) (*proto.BalanceResponse, error) {
	return w.client.Balance(ctx, &proto.BalanceRequest{Address: address})
}

// NewAddr created a new wallet address managed by the FFS instance.
func (w *Wallet) NewAddr(ctx context.Context, name string, options ...NewAddressOption) (*proto.NewAddrResponse, error) {
	r := &proto.NewAddrRequest{Name: name}
	for _, opt := range options {
		opt(r)
	}
	return w.client.NewAddr(ctx, r)
}

// Addrs returns a list of addresses managed by the FFS instance.
func (w *Wallet) Addrs(ctx context.Context) (*proto.AddrsResponse, error) {
	return w.client.Addrs(ctx, &proto.AddrsRequest{})
}

// SendFil sends fil from a managed address to any another address, returns immediately but funds are sent asynchronously.
func (w *Wallet) SendFil(ctx context.Context, from string, to string, amount int64) (*proto.SendFilResponse, error) {
	req := &proto.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return w.client.SendFil(ctx, req)
}

// SignMessage signs a message with a FFS managed wallet address.
func (w *Wallet) SignMessage(ctx context.Context, addr string, message []byte) (*proto.SignMessageResponse, error) {
	r := &proto.SignMessageRequest{Addr: addr, Msg: message}
	return w.client.SignMessage(ctx, r)
}

// VerifyMessage verifies a message signature from a wallet address.
func (w *Wallet) VerifyMessage(ctx context.Context, addr string, message, signature []byte) (*proto.VerifyMessageResponse, error) {
	r := &proto.VerifyMessageRequest{Addr: addr, Msg: message, Signature: signature}
	return w.client.VerifyMessage(ctx, r)
}
