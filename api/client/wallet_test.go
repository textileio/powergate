package client

import (
	"testing"

	pb "github.com/textileio/filecoin/wallet/pb"
)

func TestNewWallet(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	var err error
	address, err := w.NewWallet(ctx, "bls")
	if err != nil {
		t.Fatalf("failed to create new wallet: %v", err)
	}
	if len(address) < 1 {
		t.Fatal("received empty address from NewWallet")
	}
}

func TestWalletBalance(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	address, err := w.NewWallet(ctx, "bls")
	checkErr(t, err)

	bal, err := w.WalletBalance(ctx, address)
	if err != nil {
		t.Fatalf("failed to get wallet balance: %v", err)
	}
	if bal != 0 {
		t.Fatalf("unexpected wallet balance: %v", bal)
	}
}

func setupWallet(t *testing.T) (*Wallet, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Wallet{client: pb.NewAPIClient(conn)}, func() {
		done()
		serverDone()
	}
}
