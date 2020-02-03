package client

import (
	"testing"

	pb "github.com/textileio/fil-tools/index/miner/pb"
)

func TestGetMiners(t *testing.T) {
	skipIfShort(t)
	m, done := setupMiners(t)
	defer done()

	_, err := m.Get(ctx)
	if err != nil {
		t.Fatalf("failed to call Get: %v", err)
	}
}

func setupMiners(t *testing.T) (*Miners, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &Miners{client: pb.NewAPIClient(conn)}, func() {
		done()
		serverDone()
	}
}
