package chainsync

import (
	"context"
	"os"
	"testing"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/tests"
)

func TestMain(m *testing.M) {
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestPrecede(t *testing.T) {
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	ctx := context.Background()

	h, err := client.ChainHead(ctx)
	checkErr(t, err)

	csync := New(client)
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
