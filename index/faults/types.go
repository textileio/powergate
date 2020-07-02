package faults

// IndexSnapshot contains faults histoy information up-to a TipSetKey.
type IndexSnapshot struct {
	TipSetKey string
	Miners    map[string]Faults
}

// Faults contains epochs where a fault was detected for a miner.
type Faults struct {
	Epochs []int64
}

// Module provides faults information about the Filecoin network.
type Module interface {
	Get() IndexSnapshot
	Listen() <-chan struct{}
	Unregister(c chan struct{})
}
