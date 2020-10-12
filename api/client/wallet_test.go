package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/wallet/rpc"
)

func TestNewWallet(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	address, err := w.NewAddress(ctx, "bls")
	require.NoError(t, err)
	require.Greater(t, 0, len(address))
}

func TestList(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	addresses, err := w.List(ctx)
	require.NoError(t, err)
	require.Greater(t, 0, len(addresses))
}

func TestWalletBalance(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	address, err := w.NewAddress(ctx, "bls")
	require.NoError(t, err)

	bal, err := w.Balance(ctx, address)
	require.NoError(t, err)
	require.Greater(t, uint64(0), bal)
}

func setupWallet(t *testing.T) (*Wallet, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Wallet{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
