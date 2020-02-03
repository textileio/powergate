package slashing

import (
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/filecoin/tests"
	"github.com/textileio/filecoin/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 10
	logging.SetAllLoggers(logging.LevelError)
	logging.SetLogLevel("index-miner", "debug")
	os.Exit(m.Run())
}

func TestFreshIndex(t *testing.T) {
	dnet, _, _, close := tests.CreateLocalDevnet(t, 1)
	defer close()
	time.Sleep(time.Millisecond * 500) // Allow the network to some tipsets

	sh, err := New(tests.NewTxMapDatastore(), dnet.Client)
	checkErr(t, err)

	// Wait for some rounds of slashing updating
	for i := 0; i < 10; i++ {
		select {
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for miner index full refresh")
		case <-sh.Listen():
		}
	}
	index := sh.Get()
	if index.TipSetKey == "" {
		t.Fatalf("miner info state is invalid: %s %d", index.TipSetKey, len(index.Miners))
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
