package client

import (
	"testing"

	"github.com/textileio/powergate/ffs/rpc"
)

func TestCreate(t *testing.T) {
	skipIfShort(t)
	f, done := setupFfs(t)
	defer done()

	_, _, err := f.Create(ctx)
	if err != nil {
		t.Fatalf("failed to call Create: %v", err)
	}
}

func setupFfs(t *testing.T) (*ffs, func()) {
	serverDone := setupServer(t)
	conn, done := setupConnection(t)
	return &ffs{client: rpc.NewFFSClient(conn)}, func() {
		done()
		serverDone()
	}
}
