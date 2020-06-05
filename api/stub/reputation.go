package stub

import (
	"context"
	"time"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/reputation"
)

// Reputation provides an API for viewing reputation data.
type Reputation struct {
}

// AddSource adds a new external Source to be considered for reputation generation.
func (r *Reputation) AddSource(ctx context.Context, id string, maddr ma.Multiaddr) error {
	time.Sleep(time.Second * 3)
	return nil
}

// GetTopMiners gets the top n miners with best score.
func (r *Reputation) GetTopMiners(ctx context.Context, limit int) ([]reputation.MinerScore, error) {
	time.Sleep(time.Second * 3)
	topMiners := []reputation.MinerScore{
		{
			Addr:  "miner1",
			Score: 91,
		},
		{
			Addr:  "miner2",
			Score: 45,
		},
		{
			Addr:  "miner3",
			Score: 23,
		},
		{
			Addr:  "miner4",
			Score: 9,
		},
	}
	return topMiners, nil
}
