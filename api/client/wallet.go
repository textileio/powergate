package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
)

// Wallet provides an API for managing filecoin wallets.
type Wallet struct {
	walletClient    walletRpc.RPCServiceClient
	powergateClient proto.PowergateServiceClient
}

// NewAddress creates a new filecoin address [bls|secp256k1].
func (w *Wallet) NewAddress(ctx context.Context, typ string) (*walletRpc.NewAddressResponse, error) {
	return w.walletClient.NewAddress(ctx, &walletRpc.NewAddressRequest{Type: typ})
}

// List returns all wallet addresses.
func (w *Wallet) List(ctx context.Context) (*walletRpc.ListResponse, error) {
	return w.walletClient.List(ctx, &walletRpc.ListRequest{})
}

// Balance gets a filecoin wallet's balance.
func (w *Wallet) Balance(ctx context.Context, address string) (*proto.BalanceResponse, error) {
	return w.powergateClient.Balance(ctx, &proto.BalanceRequest{Address: address})
}
