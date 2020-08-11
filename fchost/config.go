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
			"/dns4/bootstrap-0.calibration.fildev.network/tcp/1347/p2p/12D3KooWPmhFGJkE7wDUdtzDYr7ReML9vgzJ8Tv7ubh9T6Le1Bmn",
			"/dns4/bootstrap-2.calibration.fildev.network/tcp/1347/p2p/12D3KooWPWUw5yEet6NWpxhxoibXFbLprG4k5PMLKLeubGBLf6nd",
			"/dns4/bootstrap-1.calibration.fildev.network/tcp/1347/p2p/12D3KooWGwv2YtXyYPrEKssttUT3TKZknPkCWKR6WVTvt9LW4hdf",
			"/dns4/bootstrap-3.calibration.fildev.network/tcp/1347/p2p/12D3KooWHgMU953YxD5skVG3RKa58TXwVL9z5ycGKrZdaFzGpouT",
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
