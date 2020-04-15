package health

import (
	"context"
	"testing"

	"github.com/textileio/powergate/net/lotus"
	"github.com/textileio/powergate/tests"
)

var (
	ctx = context.Background()
)

func TestModule(t *testing.T) {
	t.Parallel()

	client, _, _ := tests.CreateLocalDevnet(t, 1)
	lm := lotus.New(client, &tests.LrMock{})
	m := New(lm)

	t.Run("Health", func(t *testing.T) {
		status, messages, err := m.Check(ctx)
		tests.CheckErr(t, err)
		if status != Ok {
			t.Fatalf("expected Ok status but got %v", status)
		}
		if len(messages) != 0 {
			t.Fatalf("expected 0 messages but got %v", messages)
		}
	})
}
