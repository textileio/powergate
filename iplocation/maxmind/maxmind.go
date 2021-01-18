package maxmind

import (
	"fmt"
	"net"
	"sync"

	logger "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/textileio/powergate/v2/iplocation"
	"github.com/textileio/powergate/v2/util"
)

var (
	log = logger.Logger("maxmind")
)

// MaxMind is an iplocation resolver using the MaxMind database.
type MaxMind struct {
	lock sync.Mutex
	db   *geoip2.Reader
}

// New returns a new MaxMind iplocation resolver.
func New(db string) (*MaxMind, error) {
	r, err := geoip2.Open(db)
	if err != nil {
		return nil, fmt.Errorf("opening geolite2 database: %s", err)
	}
	return &MaxMind{db: r}, nil
}

// Resolve returns Location information from multiaddrs.
func (mm *MaxMind) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	for _, ma := range mas {
		ipport, err := util.TCPAddrFromMultiAddr(ma)
		if err != nil {
			log.Debugf("transforming %s to tcp addr: %s", ma, err)
			continue
		}
		strIP, _, err := net.SplitHostPort(ipport)
		if err != nil {
			log.Debugf("parsing ip/port from %s: %s", ipport, err)
			continue
		}
		ip := net.ParseIP(strIP)
		city, err := mm.db.City(ip)
		if err != nil {
			log.Debugf("querying maxmind db for %s: %s", ipport, err)
			continue
		}
		if city.Country.IsoCode != "" || (city.Location.Latitude != 0 && city.Location.Longitude != 0) {
			return iplocation.Location{
				Country:   city.Country.IsoCode,
				Latitude:  city.Location.Latitude,
				Longitude: city.Location.Longitude,
			}, nil
		}
		log.Debugf("no info for addr %s", ip)
	}
	return iplocation.Location{}, iplocation.ErrCantResolve
}

// Close closes the iplocation resolver.
func (mm *MaxMind) Close() error {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	if err := mm.db.Close(); err != nil {
		return fmt.Errorf("closing geolite2 database: %s", err)
	}
	return nil
}
