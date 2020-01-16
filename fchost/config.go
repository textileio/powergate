package fchost

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
)

var (
	addrs = []string{
		"/dns4/lotus-bootstrap-0.sin.fil-test.net/tcp/1347/p2p/12D3KooWLZs8BWtEzRTYET4yR4jzDtPamaA1YsyPQJq6cf2RfxBD",
		"/dns4/lotus-bootstrap-1.sin.fil-test.net/tcp/1347/p2p/12D3KooWGvrgjWw4Yqo4AFWqYp4g37FpUvUCQBkNWudZVSwR9tY1",
		"/dns4/lotus-bootstrap-0.fra.fil-test.net/tcp/1347/p2p/12D3KooWSfNcrD1cs5Cj5eSHbK6nHCqJLffAuPqvRMBRgvUdqQhX",
		"/dns4/lotus-bootstrap-1.fra.fil-test.net/tcp/1347/p2p/12D3KooWNkXyVPspUnrHUiSC3VJPMcXvHuNdy3BTCLTPPnDgwwTT",
		"/dns4/lotus-bootstrap-0.dfw.fil-test.net/tcp/1347/p2p/12D3KooWSgJWJZK8LTRtCWzPa5FQheCFJjHpficVYgEQWeimcqCu",
		"/dns4/lotus-bootstrap-1.dfw.fil-test.net/tcp/1347/p2p/12D3KooWFPaC4dyGpbNXCpVHjZucdJnDwmv4ng9tponPx5GrzJkT",
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
