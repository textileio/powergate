package miner

import (
	"time"
)

// IndexSnapshot contains on-chain and off-chain information about miners.
type IndexSnapshot struct {
	Meta    MetaIndex
	OnChain ChainIndex
}

// ChainIndex contains on-chain information about miners.
type ChainIndex struct {
	LastUpdated int64
	Miners      map[string]OnChainMinerData
}

// OnChainMinerData contains on-chain data about a miner.
type OnChainMinerData struct {
	Power         uint64
	RelativePower float64
	SectorSize    uint64
	SectorsLive   uint64
	SectorsActive uint64
	SectorsFaulty uint64
}

// MetaIndex contains off-chain information about miners.
type MetaIndex struct {
	Info map[string]Meta
}

// Meta contains off-chain information of a miner.
type Meta struct {
	LastUpdated time.Time
	UserAgent   string
	Location    Location
}

// Location contains geeoinformation.
type Location struct {
	Country   string
	Longitude float64
	Latitude  float64
}

// Module provides information about miners in the
// Filecoin network.
type Module interface {
	Get() IndexSnapshot
	Listen() <-chan struct{}
	Unregister(c chan struct{})
}
