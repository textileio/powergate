package types

import cbor "github.com/ipfs/go-ipld-cbor"

func init() {
	cbor.RegisterCborType(Index{})
	cbor.RegisterCborType(Slashes{})
}

// Index contains slashing histoy information up-to a TipSetKey.
type Index struct {
	TipSetKey string
	Miners    map[string]Slashes
}

// Slashes contains a slice of distinct epochs for a miner
type Slashes struct {
	Epochs []uint64
}
