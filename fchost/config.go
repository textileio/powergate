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
		"calibrationnet": {
			"/dns4/bootstrap-0.calibration.fildev.network/tcp/1347/p2p/12D3KooWDHzHtmsH9BMf9is7hCzJnd1jtVb5rdeAmfm9CCoZFqeY",
			"/dns4/bootstrap-1.calibration.fildev.network/tcp/1347/p2p/12D3KooWHyTNdCjTviTCMpFiyYVb3TamxuLmM3wEcrhH5T4M9Q3J",
			"/dns4/bootstrap-2.calibration.fildev.network/tcp/1347/p2p/12D3KooWBM8Zdq288tyYF5yT1j2ym9ynApddGJFjYnXNiFiCbdXp",
			"/dns4/bootstrap-3.calibration.fildev.network/tcp/1347/p2p/12D3KooWKoegjZRfgY8Zp6QMqr4UtuJHzTKeWVmrniFNXeikzvGT",
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
