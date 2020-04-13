package slashing

import cbor "github.com/ipfs/go-ipld-cbor"

func init() {
	cbor.RegisterCborType(IndexSnapshot{})
	cbor.RegisterCborType(Slashes{})
}

// IndexSnapshot contains slashing histoy information up-to a TipSetKey.
type IndexSnapshot struct {
	TipSetKey string
	Miners    map[string]Slashes
}

// Slashes contains a slice of distinct epochs for a miner
type Slashes struct {
	Epochs []uint64
}
