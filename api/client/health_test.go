package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	h "github.com/textileio/powergate/health"
	"github.com/textileio/powergate/health/rpc"
)

func TestCheck(t *testing.T) {
	c, done := setupHealth(t)
	defer done()
	status, messages, err := c.Check(ctx)
	require.Nil(t, err)
	require.Empty(t, messages)
	require.Equal(t, h.Ok, status)
}

func setupHealth(t *testing.T) (*health, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &health{client: rpc.NewHealthClient(conn)}, func() {
		done()
		serverDone()
	}
}
