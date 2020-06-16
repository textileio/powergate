package fchost

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	addrs = []string{
		"dns4/t01000.miner.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWGn5vqmazSwXBQnPAcQiFyXDvpngTaFu5y9EYD9XCiRXn",
		"/ip4/34.217.110.132/tcp/1347/p2p/12D3KooWGn5vqmazSwXBQnPAcQiFyXDvpngTaFu5y9EYD9XCiRXn",
		"/dns4/peer0.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWBoh25ZEWSwrbmQXANX4BH4ZczEaePFh6nNTbVQfRdfZ8",
		"/ip4/54.187.182.170/tcp/1347/p2p/12D3KooWBoh25ZEWSwrbmQXANX4BH4ZczEaePFh6nNTbVQfRdfZ8",
		"/dns4/peer1.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWLdj4UyD2ucKMbXV4hKZw13r9zQMLkYucVDxubt8S6XJ8",
		"/ip4/52.24.84.39/tcp/1347/p2p/12D3KooWLdj4UyD2ucKMbXV4hKZw13r9zQMLkYucVDxubt8S6XJ8",
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
