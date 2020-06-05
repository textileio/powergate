package stub

import (
	"context"
	"time"
)

// Wallet provides an API for managing filecoin wallets.
type Wallet struct {
}

// NewWallet creates a new filecoin wallet [bls|secp256k1].
func (w *Wallet) NewWallet(ctx context.Context, typ string) (string, error) {
	time.Sleep(time.Millisecond * 1000)
	return "t16yt7dydsey3rzhefn4tudpn22pmp5ty2rmqyx7y", nil
}

// WalletBalance gets a filecoin wallet's balance.
func (w *Wallet) WalletBalance(ctx context.Context, address string) (int64, error) {
	time.Sleep(time.Millisecond * 1000)
	return 47839, nil
}
