package ask

import (
	"time"
)

// Module provides information about storage asks.
type Module interface {
	Get() Index
	Query(q Query) ([]StorageAsk, error)
	Listen() <-chan struct{}
	Unregister(c chan struct{})
}

// Index contains Ask information from markets.
type Index struct {
	LastUpdated        time.Time
	StorageMedianPrice uint64
	Storage            map[string]StorageAsk
}

// StorageAsk has information about an active ask from a storage miner.
type StorageAsk struct {
	Miner        string
	Price        uint64
	MinPieceSize uint64
	MaxPieceSize uint64
	Timestamp    int64
	Expiry       int64
}

// Query specifies filtering and paging data to retrieve active Asks.
type Query struct {
	MaxPrice  uint64
	PieceSize uint64
	Limit     int
	Offset    int
}
