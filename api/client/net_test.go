package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	n "github.com/textileio/powergate/net"
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
	peers, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, peers)
}

func TestFindPeer(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	res0, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res0)
	res1, err := c.FindPeer(ctx, res0.Peers[0].AddrInfo.Id)
	require.NoError(t, err)
	require.NotEmpty(t, res1.PeerInfo.AddrInfo.Id)
	require.NotEmpty(t, res1.PeerInfo.AddrInfo.Addrs)
	// The addrs of peers are in localhost, so
	// no location information will be available.
	require.Nil(t, res1.PeerInfo.Location)
}

func TestDisconnectConnect(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	res, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res.Peers)
	err = c.DisconnectPeer(ctx, res.Peers[0].AddrInfo.Id)
	require.NoError(t, err)
	err = c.ConnectPeer(ctx, res.Peers[0].AddrInfo)
	require.NoError(t, err)
}

func TestConnectedness(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	res, err := c.Peers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res.Peers)
	connectedness, err := c.Connectedness(ctx, res.Peers[0].AddrInfo.Id)
	require.NoError(t, err)
	require.Equal(t, n.Connected, connectedness)
}

func setupNet(t *testing.T) (*Net, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Net{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
