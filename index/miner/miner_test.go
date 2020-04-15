package miner

import (
	"os"
	"testing"
	"time"

	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	cbor.RegisterCborType(time.Time{})
	util.AvgBlockTime = time.Millisecond * 10
	metadataRefreshInterval = time.Millisecond * 10
	logging.SetAllLoggers(logging.LevelError)
	//logging.SetLogLevel("index-miner", "debug")
	os.Exit(m.Run())
}

func TestFullRefresh(t *testing.T) {
	client, _, miners := tests.CreateLocalDevnet(t, 1)
	time.Sleep(time.Second * 15) // Allow the network to some tipsets

	mi, err := New(tests.NewTxMapDatastore(), client, &tests.P2pHostMock{}, &tests.LrMock{})
	checkErr(t, err)

	l := mi.Listen()
	// Wait for some rounds of on-chain and meta updates
	for i := 0; i < 10; i++ {
		select {
		case <-time.After(time.Second * 30):
			t.Fatal("timeout waiting for miner index full refresh")
		case <-l:
		}
	}

	index := mi.Get()
	if index.Chain.LastUpdated == 0 || len(index.Chain.Power) != len(miners) {
		t.Fatalf("miner info state is invalid: %d %d", index.Chain.LastUpdated, len(index.Chain.Power))
	}
	if index.Meta.Online != uint32(len(miners)) || index.Meta.Offline > 0 {
		t.Fatalf("meta index has wrong information")
	}
	for _, m := range miners {
		chainInfo, ok := index.Chain.Power[m.String()]
		if !ok {
			t.Fatalf("on-chain power info for miner %s is missing", m.String())
		}
		if chainInfo.Power == 0 || chainInfo.Relative == 0 {
			t.Fatalf("invalid values for miner %s power: %v", m.String(), chainInfo)
		}

		metaInfo, ok := index.Meta.Info[m.String()]
		if !ok {
			t.Fatalf("meta info for miner %s is missing", m.String())
		}
		emptyTime := time.Time{}
		if metaInfo.LastUpdated == emptyTime ||
			metaInfo.UserAgent == "" ||
			!metaInfo.Online {
			t.Fatalf("invalid meta values for miner %s: %v", m.String(), metaInfo)
		}
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
