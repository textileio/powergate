package lotus

import (
	"context"
	"testing"

	"github.com/textileio/powergate/net"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/tests/mocks"
)

var (
	ctx = context.Background()
)

func TestModule(t *testing.T) {
	t.Parallel()

	client, _, _ := tests.CreateLocalDevnet(t, 1)
	m := New(client, &mocks.LrMock{})

	t.Run("ListenAddr", func(t *testing.T) {
		addrInfo, err := m.ListenAddr(ctx)
		tests.CheckErr(t, err)
		if addrInfo.ID.String() == "" {
			t.Fatalf("got empty addrInfo.ID")
		}
		if len(addrInfo.Addrs) == 0 {
			t.Fatalf("got empty addrInfo.Addrs")
		}
	})

	t.Run("Peers", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers) == 0 {
			t.Fatalf("got empty peers list")
		}
	})

	t.Run("Connectedness", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers) == 0 {
			t.Fatalf("got empty peers list")
		}
		c, err := m.Connectedness(ctx, peers[0].AddrInfo.ID)
		tests.CheckErr(t, err)
		if c != net.Connected {
			t.Fatalf("not connected to peer")
		}
	})

	t.Run("FindPeer", func(t *testing.T) {
		peers, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers) == 0 {
			t.Fatalf("got empty peers list")
		}
		peer, err := m.FindPeer(ctx, peers[0].AddrInfo.ID)
		tests.CheckErr(t, err)
		if peer.AddrInfo.ID != peers[0].AddrInfo.ID {
			t.Fatalf("expected to find peer %v but found peer %v", peers[0].AddrInfo.ID.String(), peer.AddrInfo.ID.String())
		}
	})

	t.Run("DisconnectAndConnect", func(t *testing.T) {
		peers0, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers0) == 0 {
			t.Fatalf("got empty peers list")
		}

		err = m.DisconnectPeer(ctx, peers0[0].AddrInfo.ID)
		tests.CheckErr(t, err)
		peers1, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers1) != 0 {
			t.Fatalf("expected no peers but have peers")
		}

		err = m.ConnectPeer(ctx, peers0[0].AddrInfo)
		tests.CheckErr(t, err)
		peers2, err := m.Peers(ctx)
		tests.CheckErr(t, err)
		if len(peers2) != 1 {
			t.Fatalf("expected one peer but found %v", len(peers2))
		}
	})
}
