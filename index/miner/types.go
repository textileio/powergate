package miner

import (
	"time"

	cbor "github.com/ipfs/go-ipld-cbor"
)

func init() {
	cbor.RegisterCborType(IndexSnapshot{})
	cbor.RegisterCborType(ChainIndex{})
	cbor.RegisterCborType(Power{})
	cbor.RegisterCborType(MetaIndex{})
	cbor.RegisterCborType(Meta{})
	cbor.RegisterCborType(Location{})
}

// IndexSnapshot contains on-chain and off-chain information about miners
type IndexSnapshot struct {
	Meta  MetaIndex
	Chain ChainIndex
}

// ChainIndex contains on-chain information about miners
type ChainIndex struct {
	LastUpdated int64
	Power       map[string]Power
}

// Power contains power information of a miner
type Power struct {
	Power    uint64
	Relative float64
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
