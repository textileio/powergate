package iplocation

import (
	"errors"

	"github.com/multiformats/go-multiaddr"
)

var (
	// ErrCantResolve indicates that geoinformation couldn't be resolved for a host.
	ErrCantResolve = errors.New("can't resolve multiaddr location information")
)

// Location contains geoinformation.
type Location struct {
	Country   string
	Latitude  float64
	Longitude float64
}

// LocationResolver resolver gets location information from a set of multiaddresses of
// a single host.
type LocationResolver interface {
	Resolve(mas []multiaddr.Multiaddr) (Location, error)
}
