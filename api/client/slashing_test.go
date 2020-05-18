package client

import (
	"testing"

	"github.com/textileio/powergate/index/slashing/rpc"
)

func TestGetSlashing(t *testing.T) {
	skipIfShort(t)
	s, done := setupSlashing(t)
	defer done()

	_, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("failed to call Get: %v", err)
	}
}

func setupSlashing(t *testing.T) (*Slashing, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Slashing{client: rpc.NewAPIClient(conn)}, func() {
		done()
		serverDone()
	}
}
