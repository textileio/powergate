package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"github.com/textileio/powergate/wallet/rpc"
)

func TestNewWallet(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	res, err := w.NewAddress(ctx, "bls")
	require.NoError(t, err)
	require.Greater(t, len(res.Address), 0)
}

func TestList(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	res, err := w.List(ctx)
	require.NoError(t, err)
	require.Greater(t, len(res.Addresses), 0)
}

func TestWalletBalance(t *testing.T) {
	w, done := setupWallet(t)
	defer done()

	newAddressRes, err := w.NewAddress(ctx, "bls")
	require.NoError(t, err)

	_, err = w.Balance(ctx, newAddressRes.Address)
	require.NoError(t, err)
}

func setupWallet(t *testing.T) (*Wallet, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Wallet{walletClient: rpc.NewRPCServiceClient(conn), powergateClient: proto.NewPowergateServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
