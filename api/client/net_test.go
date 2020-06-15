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
	addrInfo, err := c.ListenAddr(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, addrInfo.Addrs)
	require.NotEmpty(t, addrInfo.ID)
}

func TestPeers(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peers, err := c.Peers(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, peers)
}

func TestFindPeer(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peers, err := c.Peers(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, peers)
	peer, err := c.FindPeer(ctx, peers[0].AddrInfo.ID)
	require.Nil(t, err)
	require.NotEmpty(t, peer.AddrInfo.ID)
	require.NotEmpty(t, peer.AddrInfo.Addrs)
	// The addrs of peers are in localhost, so
	// no location information will be available.
	require.Nil(t, peer.Location)
}

func TestDisconnectConnect(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peers, err := c.Peers(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, peers)
	err = c.DisconnectPeer(ctx, peers[0].AddrInfo.ID)
	require.Nil(t, err)
	err = c.ConnectPeer(ctx, peers[0].AddrInfo)
	require.Nil(t, err)
}

func TestConnectedness(t *testing.T) {
	c, done := setupNet(t)
	defer done()
	peers, err := c.Peers(ctx)
	require.Nil(t, err)
	require.NotEmpty(t, peers)
	connectedness, err := c.Connectedness(ctx, peers[0].AddrInfo.ID)
	require.Nil(t, err)
	require.Equal(t, n.Connected, connectedness)
}

func setupNet(t *testing.T) (*Net, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Net{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
