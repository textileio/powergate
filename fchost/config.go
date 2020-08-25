package fchost

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	networkBootstrappers = map[string][]string{
		"testnetnet": {
			"/dns4/bootstrap-0.testnet.fildev.network/tcp/1347/p2p/12D3KooWJTUBUjtzWJGWU1XSiY21CwmHaCNLNYn2E7jqHEHyZaP7",
			"/dns4/bootstrap-1.testnet.fildev.network/tcp/1347/p2p/12D3KooW9yeKXha4hdrJKq74zEo99T8DhriQdWNoojWnnQbsgB3v",
			"/dns4/bootstrap-2.testnet.fildev.network/tcp/1347/p2p/12D3KooWCrx8yVG9U9Kf7w8KLN3Edkj5ZKDhgCaeMqQbcQUoB6CT",
			"/dns4/bootstrap-4.testnet.fildev.network/tcp/1347/p2p/12D3KooWPkL9LrKRQgHtq7kn9ecNhGU9QaziG8R5tX8v9v7t3h34",
			"/dns4/bootstrap-3.testnet.fildev.network/tcp/1347/p2p/12D3KooWKYSsbpgZ3HAjax5M1BXCwXLa6gVkUARciz7uN3FNtr7T",
			"/dns4/bootstrap-5.testnet.fildev.network/tcp/1347/p2p/12D3KooWQYzqnLASJAabyMpPb1GcWZvNSe7JDcRuhdRqonFoiK9W",
		},
	}
)

func getBootstrapPeers(network string) ([]peer.AddrInfo, error) {
	addrs, ok := networkBootstrappers[network]
	if !ok {
		return nil, fmt.Errorf("network doesn't have any configured bootstrappers")
	}

	maddrs := make([]multiaddr.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("converting multiaddrs: %s", err)
		}
	}
	peers, err := peer.AddrInfosFromP2pAddrs(maddrs...)
	if err != nil {
		return nil, fmt.Errorf("multiaddr conversion: %s", err)
	}
	return peers, nil
}

func getDefaultOpts() []config.Option {
	return []config.Option{libp2p.Defaults}
}
