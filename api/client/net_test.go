package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	n "github.com/textileio/powergate/net"
	"github.com/textileio/powergate/net/rpc"
)

func TestNet(t *testing.T) {
	c, done := setupNet(t)
	defer done()

	t.Run("ListenAddr", func(t *testing.T) {
		addrInfo, err := c.ListenAddr(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, addrInfo.Addrs)
		require.NotEmpty(t, addrInfo.ID)
	})

	t.Run("Peers", func(t *testing.T) {
		peers, err := c.Peers(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, peers)
	})

	t.Run("FindPeer", func(t *testing.T) {
		peers, err := c.Peers(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, peers)
		peer, err := c.FindPeer(ctx, peers[0].AddrInfo.ID)
		require.Nil(t, err)
		require.NotEmpty(t, peer.AddrInfo.ID)
		require.NotEmpty(t, peer.AddrInfo.Addrs)
		require.NotEmpty(t, peer.Location.Country)
	})

	t.Run("DisconnectConnect", func(t *testing.T) {
		peers, err := c.Peers(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, peers)
		err = c.DisconnectPeer(ctx, peers[0].AddrInfo.ID)
		require.Nil(t, err)
		err = c.ConnectPeer(ctx, peers[0].AddrInfo)
		require.Nil(t, err)
	})

	t.Run("Connectedness", func(t *testing.T) {
		peers, err := c.Peers(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, peers)
		connectedness, err := c.Connectedness(ctx, peers[0].AddrInfo.ID)
		require.Nil(t, err)
		require.Equal(t, n.Connected, connectedness)
	})
}

func setupNet(t *testing.T) (*net, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &net{client: rpc.NewNetClient(conn)}, func() {
		done()
		serverDone()
	}
}
