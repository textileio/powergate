package client

import (
	"testing"

	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	skipIfShort(t)
	done := setupServer(t)
	defer done()

	client, err := NewClient(grpcHostAddress, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("failed to close client: %v", err)
	}
}
