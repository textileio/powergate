package slashing

import (
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 100
	logging.SetAllLoggers(logging.LevelError)
	os.Exit(m.Run())
}

func TestFreshIndex(t *testing.T) {
	// Skipped until #235 lands.
	t.SkipNow()
	client, _, _ := tests.CreateLocalDevnet(t, 1)
	time.Sleep(time.Millisecond * 500) // Allow the network to some tipsets

	sh, err := New(tests.NewTxMapDatastore(), client)
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
