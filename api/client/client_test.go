package client

import (
	"testing"

	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	skipIfShort(t)
	done := setupServer(t)
	defer done()

	ma, err := multiaddr.NewMultiaddr(grpcHostAddress)
	if err != nil {
		t.Fatalf("parsing multiaddress: %s", err)
	}
	client, err := NewClient(ma, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("failed to close client: %v", err)
	}
}
