package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
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

	id, token, err := client.Ffs.Create(context.Background())
	require.Nil(t, err)
	require.NotEmpty(t, id)
	require.NotEmpty(t, token)
}
