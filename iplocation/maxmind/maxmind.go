package maxmind

import (
	"fmt"
	"net"
	"sync"

	"github.com/multiformats/go-multiaddr"
	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/textileio/powergate/iplocation"
	"github.com/textileio/powergate/util"
	"gopkg.in/src-d/go-log.v1"
)

type MaxMind struct {
	lock sync.Mutex
	db   *geoip2.Reader
}

func New(db string) (*MaxMind, error) {
	r, err := geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		return nil, fmt.Errorf("opening geolite2 database: %s", err)
	}
	return &MaxMind{db: r}, nil
}

func (mm *MaxMind) Resolve(mas []multiaddr.Multiaddr) (iplocation.Location, error) {
	for _, ma := range mas {
		ipport, err := util.TCPAddrFromMultiAddr(ma)
		if err != nil {
			log.Debugf("error transforming %s to tcp addr: %s", ma, err)
			continue
		}
		strIP, _, err := net.SplitHostPort(ipport)
		if err != nil {
			log.Debugf("error parsing ip/port from %s: %s", ipport, err)
			continue
		}
		ip := net.ParseIP(strIP)
		city, err := mm.db.City(ip)
		if err != nil {
			log.Debugf("getting city from %s: %s", ipport, err)
		}
		if city.Country.IsoCode != "" || (city.Location.Latitude != 0 && city.Location.Longitude != 0) {
			return iplocation.Location{
				Country:   city.Country.IsoCode,
				Latitude:  city.Location.Latitude,
				Longitude: city.Location.Longitude,
			}, nil
		}
		log.Debugf("no info for tcp addr %s", ip)
	}
	return iplocation.Location{}, iplocation.ErrCantResolve

}

func (mm *MaxMind) Close() error {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	if err := mm.db.Close(); err != nil {
		return fmt.Errorf("closing geolite2 database: %s", err)
	}
	return nil
}
