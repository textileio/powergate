package client

import (
	"testing"

	"github.com/textileio/powergate/index/miner/rpc"
)

func TestGetMiners(t *testing.T) {
	skipIfShort(t)
	m, done := setupMiners(t)
	defer done()

	_, err := m.Get(ctx)
	if err != nil {
		t.Fatalf("failed to call Get: %v", err)
	}
}

func setupMiners(t *testing.T) (*Miners, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Miners{client: rpc.NewRPCClient(conn)}, func() {
		done()
		serverDone()
	}
}
