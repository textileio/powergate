package fchost

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPingBootstrapers(t *testing.T) {
	// This test is skipped since interopnet is getting reset
	// a lot and boostrap peers change very frequently.
	// We can re-enable this when the network becomes stable again.
	t.SkipNow()
	h, err := New("mainnet", false)
	require.NoError(t, err)
	err = h.Bootstrap()
	require.NoError(t, err)

	bsPeers, err := getBootstrapPeers("mainnet")
	require.NoError(t, err)
	for _, addr := range bsPeers {
		pong := h.Ping(context.Background(), addr.ID)
		if pong {
			return
		}
	}
	t.Fatalf("no bootstrap peers replied")
}
