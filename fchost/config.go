package fchost

import (
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	pnet "github.com/libp2p/go-libp2p-pnet"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	addrs = []string{
		"/dns4/lotus-bootstrap-0.dfw.fil-test.net/tcp/1347/p2p/12D3KooWHwGBSiLR5ts7KW9MgH4BMzC2iXe18kwAQ8Ee3LUd1jeR",
		"/dns4/lotus-bootstrap-1.dfw.fil-test.net/tcp/1347/p2p/12D3KooWCLFaawdhLGcSpiqg43DtZ9QzPQ6HcB8Vvyu2Cnta8UWc",
		"/dns4/lotus-bootstrap-0.fra.fil-test.net/tcp/1347/p2p/12D3KooWMmaL7eaUCF6tVAghVmgozxz4uztbuFUQv6dyFpHRarHR",
		"/dns4/lotus-bootstrap-1.fra.fil-test.net/tcp/1347/p2p/12D3KooWLLpNYoKdf9NgcWudBhXLdTcXncqAsTzozw1scMMu6nS5",
		"/dns4/lotus-bootstrap-0.sin.fil-test.net/tcp/1347/p2p/12D3KooWCNL9vXaXwNs3Bu8uRAJK4pxpCyPeM7jZLSDpJma1wrV8",
		"/dns4/lotus-bootstrap-1.sin.fil-test.net/tcp/1347/p2p/12D3KooWNGGxFda1eC5U2YKAgs4ypoFHn3Z3xHCsjmFdrCcytoxm",
	}
)
var lotusKey = "/key/swarm/psk/1.0.0/\n/base16/\n20c72388e6299c7bbc1b501fdcc8abe4f89f798e9b93b2d2bc02e3c29b6a088e"

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
	protec, err := pnet.NewProtector(strings.NewReader(lotusKey))
	if err != nil {
		panic(err)
	}
	return []config.Option{libp2p.PrivateNetwork(protec)}
}
