package chainsync

import (
	"context"
	"testing"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/filecoin/lotus"

	"github.com/textileio/filecoin/tests"
)

func TestPrecede(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping since is a short test run")
	}

	addr, token := tests.ClientConfigMA()
	c, cls, err := lotus.New(addr, token)
	checkErr(t, err)
	defer cls()
	ctx := context.Background()

	h, err := c.ChainHead(ctx)
	checkErr(t, err)

	csync := New(c)
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
