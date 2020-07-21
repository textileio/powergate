package lotus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/net"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/tests/mocks"
)

var (
	ctx = context.Background()
)

func TestModule(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	m := New(client, &mocks.LrMock{})

	t.Run("ListenAddr", func(t *testing.T) {
		addrInfo, err := m.ListenAddr(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, addrInfo.ID.String())
		require.NotEmpty(t, addrInfo.Addrs)
	})

	t.Run("Peers", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, peers)
	})

	t.Run("Connectedness", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, peers)
		c, err := m.Connectedness(ctx, peers[0].AddrInfo.ID)
		require.NoError(t, err)
		require.Equal(t, net.Connected, c)
	})

	t.Run("FindPeer", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, peers)
		peer, err := m.FindPeer(ctx, peers[0].AddrInfo.ID)
		require.NoError(t, err)
		require.Equal(t, peers[0].AddrInfo.ID, peer.AddrInfo.ID)
	})

	t.Run("DisconnectAndConnect", func(t *testing.T) {
		peers0, err := m.Peers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, peers0)

		err = m.DisconnectPeer(ctx, peers0[0].AddrInfo.ID)
		require.NoError(t, err)
		peers1, err := m.Peers(ctx)
		require.NoError(t, err)
		require.Empty(t, peers1)

		err = m.ConnectPeer(ctx, peers0[0].AddrInfo)
		require.NoError(t, err)
		peers2, err := m.Peers(ctx)
		require.NoError(t, err)
		require.Len(t, peers2, 1)
	})
}
