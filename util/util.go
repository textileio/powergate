package util

import (
	"context"
	"fmt"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	dns "github.com/multiformats/go-multiaddr-dns"
)

var (
	// AvgBlockTime is the expected duration between block in two consecutive epochs.
	// Defined at the Filecoin spec level.
	AvgBlockTime = time.Second * 30
)

// TCPAddrFromMultiAddr converts a multiaddress to a string representation of a tcp address
func TCPAddrFromMultiAddr(maddr ma.Multiaddr) (addr string, err error) {
	if maddr == nil {
		err = fmt.Errorf("invalid address")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := maddr.ValueForProtocol(ma.P_DNS4); err == nil {
		maddrs, err := dns.Resolve(ctx, maddr)
		if err != nil {
			return "", err
		}
		for _, m := range maddrs {
			if _, err := m.ValueForProtocol(ma.P_IP4); err == nil {
				maddr = m
				break
			}
		}
	}

	ip4, err := maddr.ValueForProtocol(ma.P_IP4)
	if err != nil {
		return
	}
	tcp, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return
	}
	return fmt.Sprintf("%s:%s", ip4, tcp), nil
}

// MustParseAddr returns a parsed Multiaddr, or panics if invalid.
func MustParseAddr(str string) ma.Multiaddr {
	addr, err := ma.NewMultiaddr(str)
	if err != nil {
		panic(err)
	}
	return addr
}
