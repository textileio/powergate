package util

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	ma "github.com/multiformats/go-multiaddr"
	dns "github.com/multiformats/go-multiaddr-dns"
)

var (
	// AvgBlockTime is the expected duration between block in two consecutive epochs.
	// Defined at the Filecoin spec level.
	AvgBlockTime = time.Second * 30
)

const cidUndef = "CID_UNDEF"

// TCPAddrFromMultiAddr converts a multiaddress to a string representation of a tcp address.
func TCPAddrFromMultiAddr(maddr ma.Multiaddr) (string, error) {
	if maddr == nil {
		return "", fmt.Errorf("invalid address")
	}

	var ip string
	if _, err := maddr.ValueForProtocol(ma.P_DNS4); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		maddrs, err := dns.Resolve(ctx, maddr)
		if err != nil {
			return "", fmt.Errorf("resolving dns: %s", err)
		}
		for _, m := range maddrs {
			if ip, err = getIPFromMaddr(m); err == nil {
				break
			}
		}
	} else {
		ip, err = getIPFromMaddr(maddr)
		if err != nil {
			return "", fmt.Errorf("getting ip from maddr: %s", err)
		}
	}

	tcp, err := maddr.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return "", fmt.Errorf("getting port from maddr: %s", err)
	}
	return fmt.Sprintf("%s:%s", ip, tcp), nil
}

func getIPFromMaddr(maddr ma.Multiaddr) (string, error) {
	if ip, err := maddr.ValueForProtocol(ma.P_IP4); err == nil {
		return ip, nil
	}
	if ip, err := maddr.ValueForProtocol(ma.P_IP6); err == nil {
		return fmt.Sprintf("[%s]", ip), nil
	}
	return "", fmt.Errorf("no ip in multiaddr")
}

// MustParseAddr returns a parsed Multiaddr, or panics if invalid.
func MustParseAddr(str string) ma.Multiaddr {
	addr, err := ma.NewMultiaddr(str)
	if err != nil {
		panic(err)
	}
	return addr
}

// CidToString converts a cid to string, representing cid.Undef as an empty string.
func CidToString(c cid.Cid) string {
	if c == cid.Undef {
		return cidUndef
	}
	return c.String()
}

// CidFromString converts a string to a cid assuming that an empty string is cid.Undef.
func CidFromString(c string) (cid.Cid, error) {
	if c == "b" || c == cidUndef {
		return cid.Undef, nil
	}
	return cid.Decode(c)
}
