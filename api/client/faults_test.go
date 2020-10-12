package client

import (
	"testing"

	"github.com/textileio/powergate/index/faults/rpc"
)

func TestGetFaults(t *testing.T) {
	s, done := setupFaults(t)
	defer done()

	_, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("failed to call Get: %v", err)
	}
}

func setupFaults(t *testing.T) (*Faults, func()) {
	serverDone := setupServer(t, defaultServerConfig(t))
	conn, done := setupConnection(t)
	return &Faults{client: rpc.NewRPCServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
