package miner

import (
	"context"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/iplocation"
	"github.com/textileio/powergate/tests"
	"github.com/textileio/powergate/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 10
	metadataRefreshInterval = time.Millisecond * 10
	logging.SetAllLoggers(logging.LevelError)
	//logging.SetLogLevel("index-miner", "debug")
	os.Exit(m.Run())
}

func TestFullRefresh(t *testing.T) {
	client, _, miners := tests.CreateLocalDevnet(t, 1)
	time.Sleep(time.Second * 15) // Allow the network to some tipsets

	mi, err := New(tests.NewTxMapDatastore(), client, &p2pHostMock{}, &lrMock{})
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
	if index.OnChain.LastUpdated == 0 || len(index.OnChain.Miners) != len(miners) {
		t.Fatalf("miner info state is invalid: %d %d", index.OnChain.LastUpdated, len(index.OnChain.Miners))
	}
	if index.Meta.Online != uint32(len(miners)) || index.Meta.Offline > 0 {
		t.Fatalf("meta index has wrong information")
	}
	for _, m := range miners {
		chainInfo, ok := index.OnChain.Miners[m.String()]
		if !ok {
			t.Fatalf("on-chain power info for miner %s is missing", m.String())
		}
		if chainInfo.Power == 0 || chainInfo.RelativePower == 0 {
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

var _ P2PHost = (*p2pHostMock)(nil)

type p2pHostMock struct{}

func (hm *p2pHostMock) Addrs(id peer.ID) []multiaddr.Multiaddr {
	return nil
}
func (hm *p2pHostMock) GetAgentVersion(id peer.ID) string {
	return "fakeAgentVersion"
}
func (hm *p2pHostMock) Ping(ctx context.Context, pid peer.ID) bool {
	return true
}

var _ iplocation.LocationResolver = (*lrMock)(nil)

type lrMock struct{}

func (lr *lrMock) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	return iplocation.Location{
		Country:   "USA",
		Latitude:  0.1,
		Longitude: 0.1,
	}, nil
}
