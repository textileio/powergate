package chainsync

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestPrecede(t *testing.T) {
	clientBuilder, _, _ := tests.CreateLocalDevnet(t, 1, 300)
	time.Sleep(time.Second * 5) // Give time for at least 1 block to be mined.
	ctx := context.Background()
	c, cls, err := clientBuilder(context.Background())
	require.NoError(t, err)
	defer cls()

	h, err := c.ChainHead(ctx)
	checkErr(t, err)

	csync := New(clientBuilder)
	head := h.Key()
	prevhead := types.NewTipSetKey(h.Blocks()[0].Parents...)
	yes, err := csync.Precedes(ctx, prevhead, head)
	checkErr(t, err)
	if !yes {
		t.Fatal("parent of head should precedes head")
	}

	yes, err = csync.Precedes(ctx, head, prevhead)
	checkErr(t, err)
	if yes {
		t.Fatal("head shouldn't preced parent of head")
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
