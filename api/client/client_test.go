package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	done := setupServer(t, defaultServerConfig(t))
	defer done()

	client, err := NewClient(grpcHostAddress)
	require.NoError(t, err)
	err = client.Close()
	require.NoError(t, err)
}
