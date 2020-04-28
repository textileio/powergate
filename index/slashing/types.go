package slashing

// IndexSnapshot contains slashing histoy information up-to a TipSetKey.
type IndexSnapshot struct {
	TipSetKey string
	Miners    map[string]Slashes
}

// Slashes contains a slice of distinct epochs for a miner
type Slashes struct {
	Epochs []uint64
}
