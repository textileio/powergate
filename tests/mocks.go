package tests

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/index/miner"
	"github.com/textileio/powergate/iplocation"
)

var _ miner.P2PHost = (*P2pHostMock)(nil)

type P2pHostMock struct{}

func (hm *P2pHostMock) Addrs(id peer.ID) []multiaddr.Multiaddr {
	return nil
}
func (hm *P2pHostMock) GetAgentVersion(id peer.ID) string {
	return "fakeAgentVersion"
}
func (hm *P2pHostMock) Ping(ctx context.Context, pid peer.ID) bool {
	return true
}

var _ iplocation.LocationResolver = (*LrMock)(nil)

type LrMock struct{}

func (lr *LrMock) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	return iplocation.Location{
		Country:   "USA",
		Latitude:  0.1,
		Longitude: 0.1,
	}, nil
}
