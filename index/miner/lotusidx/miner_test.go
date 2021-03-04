package lotusidx

import (
	"context"
	"os"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/iplocation"
	"github.com/textileio/powergate/v2/lotus"
	"github.com/textileio/powergate/v2/tests"
	"github.com/textileio/powergate/v2/util"
)

func TestMain(m *testing.M) {
	util.AvgBlockTime = time.Millisecond * 10
	metaRefreshInterval = time.Millisecond * 10
	logging.SetAllLoggers(logging.LevelInfo)
	//logging.SetLogLevel("index-miner", "debug")
	os.Exit(m.Run())
}

func TestFullRefresh(t *testing.T) {
	client, _, miners := tests.CreateLocalDevnet(t, 1, 300)
	time.Sleep(time.Second * 15) // Allow the network to some tipsets

	mi, err := New(tests.NewTxMapDatastore(), client, &p2pHostMock{}, &lrMock{}, false, false)
	require.NoError(t, err)

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
	require.Greater(t, index.OnChain.LastUpdated, int64(0))
	require.Equal(t, len(miners), len(index.OnChain.Miners))
	for _, m := range miners {
		chainInfo, ok := index.OnChain.Miners[m.String()]
		require.True(t, ok)
		require.False(t, chainInfo.Power == 0 || chainInfo.RelativePower == 0)

		metaInfo, ok := index.Meta.Info[m.String()]
		require.True(t, ok)

		emptyTime := time.Time{}
		require.False(t, metaInfo.LastUpdated == emptyTime || metaInfo.UserAgent == "")
	}
}

func TestIntegration(t *testing.T) {
	t.SkipNow()
	metaRefreshInterval = time.Hour
	minersRefreshInterval = time.Second

	lotusHost, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5555")
	require.NoError(t, err)
	lotusToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.4KpuySIvV4n6kBEXQOle-hi1Ec3lyUmRYCknz4NQyLM"

	cb, err := lotus.NewBuilder(lotusHost, lotusToken, 1)
	require.NoError(t, err)

	mi, err := New(tests.NewTxMapDatastore(), cb, &p2pHostMock{}, &lrMock{}, false, false)
	require.NoError(t, err)

	<-time.After(time.Second * 15)
	_ = mi.Close()
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
