package client

import (
	"testing"

	pb "github.com/textileio/filecoin/wallet/pb"
)

var (
	address string
)

func TestNewWallet(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	var err error
	address, err = w.NewWallet(ctx, "<a type>")
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

	_, err := w.WalletBalance(ctx, address)
	if err != nil {
		t.Fatalf("failed to get wallet balance: %v", err)
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
