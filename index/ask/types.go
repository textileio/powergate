package ask

import (
	"time"
)

// IndexSnapshot contains Ask information from markets
type IndexSnapshot struct {
	LastUpdated        time.Time
	StorageMedianPrice uint64
	Storage            map[string]StorageAsk
}

// StorageAsk has information about an active ask from a storage miner
type StorageAsk struct {
	Miner        string
	Price        uint64
	MinPieceSize uint64
	Timestamp    int64
	Expiry       int64
}
