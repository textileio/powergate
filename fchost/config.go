package fchost

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	addrs = []string{
		"/dns4/t01000.miner.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWLVgeFhXA7ztuSS1uvk4NVzZwRZLxg8PBy1mJuHfc8J6E",
		"/ip4/34.217.110.132/tcp/1347/p2p/12D3KooWLVgeFhXA7ztuSS1uvk4NVzZwRZLxg8PBy1mJuHfc8J6E",
		"/dns4/peer0.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWBkXWpT1hLHXz7VzR8ZHCJ42P6c9Nb9532ATSguBpLTKJ",
		"/ip4/54.187.182.170/tcp/1347/p2p/12D3KooWBkXWpT1hLHXz7VzR8ZHCJ42P6c9Nb9532ATSguBpLTKJ",
		"/dns4/peer1.interopnet.kittyhawk.wtf/tcp/1347/p2p/12D3KooWJk5hQHNdCsidtX4jhKQMBXKJdo2kjhrNHUZBPmCZKwKt",
		"/ip4/52.24.84.39/tcp/1347/p2p/12D3KooWJk5hQHNdCsidtX4jhKQMBXKJdo2kjhrNHUZBPmCZKwKt",
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
