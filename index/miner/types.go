package miner

import (
	"time"
)

// IndexSnapshot contains on-chain and off-chain information about miners
type IndexSnapshot struct {
	Meta    MetaIndex
	OnChain ChainIndex
}

// ChainIndex contains on-chain information about miners
type ChainIndex struct {
	LastUpdated int64
	Miners      map[string]OnChainData
}

type OnChainData struct {
	Power         uint64
	RelativePower float64
	SectorSize    uint64
	ActiveDeals   uint64
}

// MetaIndex contains off-chain information about miners
type MetaIndex struct {
	Online  uint32
	Offline uint32
	Info    map[string]Meta
}

// Meta contains off-chain information of a miner
type Meta struct {
	LastUpdated time.Time
	UserAgent   string
	Location    Location
	Online      bool
}

// Location contains geeoinformation
type Location struct {
	Country   string
	Longitude float32
	Latitude  float32
}
