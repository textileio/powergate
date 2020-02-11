package stub

import (
	"context"
	"time"

	"github.com/textileio/fil-tools/index/slashing/types"
)

// Slashing provides an API for viewing slashing data
type Slashing struct {
}

// Get returns the current index of miner slashes data
func (s *Slashing) Get(ctx context.Context) (*types.Index, error) {
	time.Sleep(time.Second * 3)
	miners := map[string]types.Slashes{
		"miner1": types.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner2": types.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner3": types.Slashes{Epochs: []uint64{123, 234, 345, 456}},
		"miner4": types.Slashes{Epochs: []uint64{123, 234, 345, 456}},
	}

	index := &types.Index{
		TipSetKey: "kjandf7843hrnfuiq3f87q34nfq3894",
		Miners:    miners,
	}

	return index, nil
}
