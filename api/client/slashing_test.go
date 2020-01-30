package client

import (
	"testing"

	pb "github.com/textileio/filecoin/index/slashing/pb"
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
	return &Slashing{client: pb.NewAPIClient(conn)}, func() {
		done()
		serverDone()
	}
}
