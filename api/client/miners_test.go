package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/index/miner/rpc"
)

func TestGetMiners(t *testing.T) {
	m, done := setupMiners(t)
	defer done()

	_, err := m.Get(ctx)
	require.NoError(t, err)
}

func setupMiners(t *testing.T) (*Miners, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Miners{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
