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
		"testnet": {
			"/dns4/bootstrap-0-sin.fil-test.net/tcp/1347/p2p/12D3KooWPdUquftaQvoQEtEdsRBAhwD6jopbF2oweVTzR59VbHEd",
			"/ip4/86.109.15.57/tcp/1347/p2p/12D3KooWPdUquftaQvoQEtEdsRBAhwD6jopbF2oweVTzR59VbHEd",
			"/dns4/bootstrap-0-dfw.fil-test.net/tcp/1347/p2p/12D3KooWQSCkHCzosEyrh8FgYfLejKgEPM5VB6qWzZE3yDAuXn8d",
			"/ip4/139.178.84.45/tcp/1347/p2p/12D3KooWQSCkHCzosEyrh8FgYfLejKgEPM5VB6qWzZE3yDAuXn8d",
			"/dns4/bootstrap-0-fra.fil-test.net/tcp/1347/p2p/12D3KooWEXN2eQmoyqnNjde9PBAQfQLHN67jcEdWU6JougWrgXJK",
			"/ip4/136.144.49.17/tcp/1347/p2p/12D3KooWEXN2eQmoyqnNjde9PBAQfQLHN67jcEdWU6JougWrgXJK",
			"/dns4/bootstrap-1-sin.fil-test.net/tcp/1347/p2p/12D3KooWLmJkZd33mJhjg5RrpJ6NFep9SNLXWc4uVngV4TXKwzYw",
			"/ip4/86.109.15.123/tcp/1347/p2p/12D3KooWLmJkZd33mJhjg5RrpJ6NFep9SNLXWc4uVngV4TXKwzYw",
			"/dns4/bootstrap-1-dfw.fil-test.net/tcp/1347/p2p/12D3KooWGXLHjiz6pTRu7x2pkgTVCoxcCiVxcNLpMnWcJ3JiNEy5",
			"/ip4/139.178.86.3/tcp/1347/p2p/12D3KooWGXLHjiz6pTRu7x2pkgTVCoxcCiVxcNLpMnWcJ3JiNEy5",
			"/dns4/bootstrap-1-fra.fil-test.net/tcp/1347/p2p/12D3KooW9szZmKttS9A1FafH3Zc2pxKwwmvCWCGKkRP4KmbhhC4R",
			"/ip4/136.144.49.131/tcp/1347/p2p/12D3KooW9szZmKttS9A1FafH3Zc2pxKwwmvCWCGKkRP4KmbhhC4R",
		},
		"nerpanet": {
			"/dns4/bootstrap-0.nerpa.fildev.network/tcp/1347/p2p/12D3KooWSTq4K1mTMoUSsDciFqxTwq3ZLu7T9scXiXZgM4tZdZi5",
			"/dns4/bootstrap-1.nerpa.fildev.network/tcp/1347/p2p/12D3KooWPWo8JEoVmekqzvAh9gsiF6hrrDYvRpgsQjqLPV4AwvZA",
			"/dns4/bootstrap-2.nerpa.fildev.network/tcp/1347/p2p/12D3KooWBMJQooyJVRdMEorxGozRdq5RLxjfHPmMrRnCcX34t7pK",
			"/dns4/bootstrap-3.nerpa.fildev.network/tcp/1347/p2p/12D3KooW9qdXNp4x51GHnUsaFJzWXcHFhp4t2HeEZNFBZj5JHfbU",
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
