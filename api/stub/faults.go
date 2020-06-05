package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/faults"
)

// Faults provides an API for viewing faults data.
type Faults struct {
}

// Get returns the current index of miner faults data.
func (s *Faults) Get(ctx context.Context) (*faults.IndexSnapshot, error) {
	time.Sleep(time.Second * 3)
	miners := map[string]faults.Faults{
		"miner1": {Epochs: []int64{123, 234, 345, 456}},
		"miner2": {Epochs: []int64{123, 234, 345, 456}},
		"miner3": {Epochs: []int64{123, 234, 345, 456}},
		"miner4": {Epochs: []int64{123, 234, 345, 456}},
	}

	index := &faults.IndexSnapshot{
		TipSetKey: "kjandf7843hrnfuiq3f87q34nfq3894",
		Miners:    miners,
	}

	return index, nil
}
