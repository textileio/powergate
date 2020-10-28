package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/api/client/utils"
)

func TestClient(t *testing.T) {
	done := utils.SetupServer(t, utils.DefaultServerConfig(t))
	defer done()

	client, err := NewClient("127.0.0.1:5002")
	require.NoError(t, err)
	err = client.Close()
	require.NoError(t, err)
}
