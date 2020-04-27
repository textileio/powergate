package fchost

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	addrs = []string{
		"/dns4/t01000.miner.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWEbScoeDSJtaEjXEwqMHYBghkrmct487mm4H3pG2fie3t",
		"/ip4/54.186.82.90/tcp/1347/p2p/12D3KooWKqBJPXy1Ei7Ww6o4BuXU3bs7mANi4VKbPiGqc1CRX1A2",
	}
)

func getBootstrapPeers() []peer.AddrInfo {
	maddrs := make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = multiaddr.NewMultiaddr(addr)
		if err != nil {
			panic(err)
		}
	}
	peers, err := peer.AddrInfosFromP2pAddrs(maddrs...)
	if err != nil {
		panic(err)
	}
	return peers
}

func getDefaultOpts() []config.Option {
	return []config.Option{libp2p.Defaults}
}
