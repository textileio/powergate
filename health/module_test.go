package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/net/lotus"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/tests/mocks"
)

var (
	ctx = context.Background()
)

func TestModule(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	lm := lotus.New(client, &mocks.LrMock{})
	m := New(lm)

	t.Run("Health", func(t *testing.T) {
		status, messages, err := m.Check(ctx)
		require.Nil(t, err)
		require.Equal(t, Ok, status)
		require.Empty(t, messages)
	})
}
