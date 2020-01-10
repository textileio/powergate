package ip2location

import (
	"sync"

	"github.com/ip2location/ip2location-go"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/filecoin/iplocation"
	"github.com/textileio/filecoin/util"
)

// Need of using package globals since the library has global state initialization.
// Cover a bit of double initialization.
var (
	lock     sync.Mutex
	instance *IP2Location
)

var _ iplocation.LocationResolver = (*IP2Location)(nil)

type IP2Location struct {
}

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

func (il *IP2Location) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	for _, ma := range mas {
		ip, err := util.TCPAddrFromMultiAddr(ma)
		if err != nil {
			continue
		}
		r := ip2location.Get_all(ip)
		if r.Country_long != "" && r.Latitude != 0 && r.Longitude != 0 {
			return iplocation.Location{
				Country:   r.Country_long,
				Latitude:  r.Latitude,
				Longitude: r.Longitude,
			}, nil
		}
	}
	return iplocation.Location{}, iplocation.ErrCantResolve
}

func (il *IP2Location) Close() {
	lock.Lock()
	defer lock.Unlock()
	instance = nil
	ip2location.Close()
}
