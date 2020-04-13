package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/slashing"
)

// Slashing provides an API for viewing slashing data
type Slashing struct {
}

// Get returns the current index of miner slashes data
func (s *Slashing) Get(ctx context.Context) (*slashing.IndexSnapshot, error) {
	time.Sleep(time.Second * 3)
	miners := map[string]slashing.Slashes{
		"miner1": slashing.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner2": slashing.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner3": slashing.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner4": slashing.Slashes{Epochs: []uint64{123, 234, 345, 456}},
	}

	index := &slashing.IndexSnapshot{
		TipSetKey: "kjandf7843hrnfuiq3f87q34nfq3894",
		Miners:    miners,
	}

	return index, nil
}
