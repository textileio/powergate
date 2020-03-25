package ask

import (
	"time"

	cbor "github.com/ipfs/go-ipld-cbor"
)

func init() {
	cbor.RegisterCborType(Index{})
	cbor.RegisterCborType(StorageAsk{})
	cbor.RegisterCborType(time.Time{})
}

// Index contains Ask information from markets
type Index struct {
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
	ChainEpoch   int64
	Expiry       int64
}
