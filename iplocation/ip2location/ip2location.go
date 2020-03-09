package ip2location

import (
	"net"
	"sync"

	"github.com/ip2location/ip2location-go"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/iplocation"
	"github.com/textileio/powergate/util"
)

// Need of using package globals since the library has global state initialization.
// Cover a bit of double initialization.
var (
	lock     sync.Mutex
	instance *IP2Location

	log = logging.Logger("ip2location")
)

var _ iplocation.LocationResolver = (*IP2Location)(nil)

// IP2Location is a LocationResolver implementation to get geoinformation of
// multiaddrs of a host
type IP2Location struct {
}

// New returns a new IP2Location
func New(paths []string) *IP2Location {
	lock.Lock()
	defer lock.Unlock()
	if instance != nil {
		return instance
	}

	for _, path := range paths {
		ip2location.Open(path)
	}
	instance = &IP2Location{}
	return instance
}

// Resolve returns geoinformation about a set of multiaddresses of a single host
func (il *IP2Location) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	for _, ma := range mas {
		ipport, err := util.TCPAddrFromMultiAddr(ma)
		if err != nil {
			log.Debugf("error transforming %s to tcp addr: %s", ma, err)
			continue
		}
		ip, _, err := net.SplitHostPort(ipport)
		if err != nil {
			log.Debugf("error parsing ip/port from %s: %s", ipport, err)
			continue
		}
		r := ip2location.Get_all(ip)
		if r.Country_long != "" || (r.Latitude != 0 && r.Longitude != 0) {
			return iplocation.Location{
				Country:   r.Country_long,
				Latitude:  r.Latitude,
				Longitude: r.Longitude,
			}, nil
		}
		log.Debugf("no info for tcp addr %s", ip)
	}
	return iplocation.Location{}, iplocation.ErrCantResolve
}

// Close closes IPLocation resolver
func (il *IP2Location) Close() {
	lock.Lock()
	defer lock.Unlock()
	instance = nil
	ip2location.Close()
}
