package lotus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/tests"
)

var (
	ctx = context.Background()
)

func TestModule(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	m := New(client)

	t.Run("List", func(t *testing.T) {
		paychInfos, err := m.List(ctx)
		require.Nil(t, err)
		require.Empty(t, paychInfos)
	})

	t.Run("Create", func(t *testing.T) {
		// ToDo
	})

	t.Run("Redeem", func(t *testing.T) {
		// ToDo
	})
}
