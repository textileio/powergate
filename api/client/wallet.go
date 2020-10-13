package client

import (
	"context"
	"fmt"

	proto "github.com/textileio/powergate/proto/powergate/v1"
	walletRpc "github.com/textileio/powergate/wallet/rpc"
)

// Wallet provides an API for managing filecoin wallets.
type Wallet struct {
	walletClient    walletRpc.RPCServiceClient
	powergateClient proto.PowergateServiceClient
}

// NewAddress creates a new filecoin address [bls|secp256k1].
func (w *Wallet) NewAddress(ctx context.Context, typ string) (string, error) {
	resp, err := w.walletClient.NewAddress(ctx, &walletRpc.NewAddressRequest{Type: typ})
	if err != nil {
		return "", fmt.Errorf("calling NewAddress: %v", err)
	}
	return resp.GetAddress(), nil
}

// List returns all wallet addresses.
func (w *Wallet) List(ctx context.Context) ([]string, error) {
	resp, err := w.walletClient.List(ctx, &walletRpc.ListRequest{})
	if err != nil {
		return nil, fmt.Errorf("calling List: %v", err)
	}
	return resp.Addresses, nil
}

// Balance gets a filecoin wallet's balance.
func (w *Wallet) Balance(ctx context.Context, address string) (uint64, error) {
	resp, err := w.powergateClient.Balance(ctx, &proto.BalanceRequest{Address: address})
	if err != nil {
		return 0, fmt.Errorf("calling Balance: %v", err)
	}
	return resp.GetBalance(), nil
}

// SendFil sends Fils from one address to another.
func (w *Wallet) SendFil(ctx context.Context, from, to string, amount int64) error {
	req := &walletRpc.SendFilRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}
	_, err := w.walletClient.SendFil(ctx, req)
	if err != nil {
		return fmt.Errorf("calling SendFil: %v", err)
	}
	return nil
}
