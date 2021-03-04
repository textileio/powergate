package mocks

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	minerModule "github.com/textileio/powergate/v2/index/miner/lotusidx"
	"github.com/textileio/powergate/v2/iplocation"
)

var _ minerModule.P2PHost = (*P2pHostMock)(nil)

// P2pHostMock provides a mock P2PHost.
type P2pHostMock struct{}

// Addrs implements Addrs.
func (hm *P2pHostMock) Addrs(id peer.ID) []multiaddr.Multiaddr {
	return nil
}

// GetAgentVersion implements GetAgentVersion.
func (hm *P2pHostMock) GetAgentVersion(id peer.ID) string {
	return "fakeAgentVersion"
}

// Ping implements Ping.
func (hm *P2pHostMock) Ping(ctx context.Context, pid peer.ID) bool {
	return true
}

var _ iplocation.LocationResolver = (*LrMock)(nil)

// LrMock provides a mock LocationResolver.
type LrMock struct{}

// Resolve implements Resolve.
func (lr *LrMock) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	return iplocation.Location{
		Country:   "USA",
		Latitude:  0.1,
		Longitude: 0.1,
	}, nil
}
