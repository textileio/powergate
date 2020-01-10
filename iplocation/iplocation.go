package iplocation

import (
	"errors"

	"github.com/multiformats/go-multiaddr"
)

var (
	ErrCantResolve = errors.New("can't resolve multiaddr location information")
)

type Location struct {
	Country   string
	Latitude  float32
	Longitude float32
}

type LocationResolver interface {
	Resolve(mas []multiaddr.Multiaddr) (Location, error)
}
