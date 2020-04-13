package ask

import (
	"time"

	cbor "github.com/ipfs/go-ipld-cbor"
)

func init() {
	cbor.RegisterCborType(IndexSnapshot{})
	cbor.RegisterCborType(StorageAsk{})
	cbor.RegisterCborType(time.Time{})
}

// IndexSnapshot contains Ask information from markets
type IndexSnapshot struct {
	LastUpdated        time.Time
	StorageMedianPrice uint64
	Storage            map[string]StorageAsk
}

// Query specifies filtering and paging data to retrieve active Asks
type Query struct {
	MaxPrice  uint64
	PieceSize uint64
	Limit     int
	Offset    int
}

// StorageAsk has information about an active ask from a storage miner
type StorageAsk struct {
	Miner        string
	Price        uint64
	MinPieceSize uint64
	Timestamp    int64
	Expiry       int64
}
