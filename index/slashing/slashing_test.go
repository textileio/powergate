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
	addr, token := tests.ClientConfig()
	c, cls, err := lotus.New(addr, token)
	checkErr(t, err)
	defer cls()

	sh := New(c, tests.NewTxMapDatastore())
	select {
	case <-time.After(time.Second * 60):
		t.Fatal("timeout waiting for miner index full refresh")
	case <-sh.Listen():
	}
	history := sh.ConsensusHistory()
	if history.LastUpdated == 0 || len(history.History) == 0 {
		t.Fatalf("miner info state is invalid")
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
