package slashing

import (
	"testing"
	"time"

	"github.com/textileio/filecoin/lotus"
	"github.com/textileio/filecoin/tests"
)

func TestFreshIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping since is a short test run")
	}
	addr, token := tests.ClientConfigMA()
	c, cls, err := lotus.New(addr, token)
	checkErr(t, err)
	defer cls()

	sh, err := New(tests.NewTxMapDatastore(), c)
	checkErr(t, err)
	select {
	case <-time.After(time.Second * 240):
		t.Fatal("timeout waiting for miner index full refresh")
	case <-sh.Listen():
	}
	index := sh.Get()
	if index.TipSetKey != "" || len(index.Miners) == 0 {
		t.Fatalf("miner info state is invalid")
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
