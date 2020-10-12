package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/index/ask"
	"github.com/textileio/powergate/index/ask/rpc"
)

func TestGetAsks(t *testing.T) {
	a, done := setupAsks(t)
	defer done()

	_, err := a.Get(ctx)
	require.NoError(t, err)
}

func TestQuery(t *testing.T) {
	a, done := setupAsks(t)
	defer done()

	_, err := a.Query(ctx, ask.Query{MaxPrice: 5})
	require.NoError(t, err)
}

func setupAsks(t *testing.T) (*Asks, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Asks{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
