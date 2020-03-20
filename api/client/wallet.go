package client

import (
	"context"

	pb "github.com/textileio/powergate/wallet/pb"
)

// Wallet provides an API for managing filecoin wallets
type Wallet struct {
	client pb.APIClient
}

// NewWallet creates a new filecoin address [bls|secp256k1]
func (w *Wallet) NewWallet(ctx context.Context, typ string) (string, error) {
	resp, err := w.client.NewAddress(ctx, &pb.NewAddressRequest{Type: typ})
	if err != nil {
		return "", err
	}
	return resp.GetAddress(), nil
}

// WalletBalance gets a filecoin wallet's balance
func (w *Wallet) WalletBalance(ctx context.Context, address string) (uint64, error) {
	resp, err := w.client.WalletBalance(ctx, &pb.WalletBalanceRequest{Address: address})
	if err != nil {
		return 0, err
	}
	return resp.GetBalance(), nil
}
