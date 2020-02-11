package stub

import (
	"context"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/fil-tools/reputation/types"
)

// Reputation provides an API for viewing reputation data
type Reputation struct {
}

// AddSource adds a new external Source to be considered for reputation generation
func (r *Reputation) AddSource(ctx context.Context, id string, maddr ma.Multiaddr) error {
	time.Sleep(time.Second * 3)
	return nil
}

// GetTopMiners gets the top n miners with best score
func (r *Reputation) GetTopMiners(ctx context.Context, limit int) ([]types.MinerScore, error) {
	time.Sleep(time.Second * 3)
	topMiners := []types.MinerScore{
		types.MinerScore{
			Addr:  "miner1",
			Score: 91,
		},
		types.MinerScore{
			Addr:  "miner2",
			Score: 45,
		},
		types.MinerScore{
			Addr:  "miner3",
			Score: 23,
		},
		types.MinerScore{
			Addr:  "miner4",
			Score: 9,
		},
	}
	return topMiners, nil
}
