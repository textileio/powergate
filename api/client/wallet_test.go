package client

import (
	"testing"

	"github.com/textileio/powergate/wallet/rpc"
)

func TestNewWallet(t *testing.T) {
	skipIfShort(t)
	w, done := setupWallet(t)
	defer done()

	var err error
	address, err := w.NewAddress(ctx, "bls")
	if err != nil {
		t.Fatalf("failed to create new wallet: %v", err)
	}
	if len(address) < 1 {
		t.Fatal("received empty address from NewWallet")
	}
}

func TestList(t *testing.T) {
	skipIfShort(t)
	w, done := setupWallet(t)
	defer done()

	var err error
	addresses, err := w.List(ctx)
	if err != nil {
		t.Fatalf("failed to list addresses: %v", err)
	}
	if len(addresses) < 1 {
		t.Fatal("received empty addresses list")
	}
}

func TestWalletBalance(t *testing.T) {
	skipIfShort(t)
	w, done := setupWallet(t)
	defer done()

	address, err := w.NewAddress(ctx, "bls")
	checkErr(t, err)

	bal, err := w.Balance(ctx, address)
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
	return &Wallet{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
