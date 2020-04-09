package fchost

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPingBootstrapers(t *testing.T) {
	h, err := New()
	require.NoError(t, err)
	err = h.Bootstrap()
	require.NoError(t, err)

	bsPeers := getBootstrapPeers()
	for _, addr := range bsPeers {
		pong := h.Ping(context.Background(), addr.ID)
		if pong {
			return
		}
	}
	t.Fatalf("no bootstrap peers replied")
}
