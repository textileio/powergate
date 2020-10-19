package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/net/rpc"
)

func TestListenAddr(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	res, err := c.ListenAddr(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res.AddrInfo.Addrs)
	require.NotEmpty(t, res.AddrInfo.Id)
}

func TestPeers(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	res, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func TestFindPeer(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peersRes, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, peersRes.Peers)
	peerRes, err := c.FindPeer(ctx, peersRes.Peers[0].AddrInfo.Id)
	require.NoError(t, err)
	require.NotEmpty(t, peerRes.PeerInfo.AddrInfo.Id)
	require.NotEmpty(t, peerRes.PeerInfo.AddrInfo.Addrs)
	// The addrs of peers are in localhost, so
	// no location information will be available.
	require.Nil(t, peerRes.PeerInfo.Location)
}

func TestConnectedness(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peersRes, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, peersRes.Peers)
	connectednessRes, err := c.Connectedness(ctx, peersRes.Peers[0].AddrInfo.Id)
	require.NoError(t, err)
	require.Equal(t, rpc.Connectedness_CONNECTEDNESS_CONNECTED, connectednessRes.Connectedness)
}

func setupNet(t *testing.T) (*Net, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Net{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
