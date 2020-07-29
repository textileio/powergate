package client

import (
	"testing"
)

func TestClient(t *testing.T) {
	skipIfShort(t)
	done := setupServer(t)
	defer done()

	client, err := NewClient(grpcHostAddress)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("failed to close client: %v", err)
	}
}
