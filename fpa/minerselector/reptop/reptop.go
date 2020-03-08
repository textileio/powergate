package reptop

import (
	"fmt"

	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/index/ask"
	"github.com/textileio/fil-tools/reputation"
)

// RepTop is a fpa.MinerSelector implementation that returns the top N
// miners from a Reputations Module and an Ask Index.
type RepTop struct {
	rm *reputation.Module
	ai *ask.AskIndex
}

var _ fpa.MinerSelector = (*RepTop)(nil)

// New returns a new RetTop instance that uses the specified Reputation Module
// to select miners and the AskIndex for their epoch prices.
func New(rm *reputation.Module, ai *ask.AskIndex) *RepTop {
	return &RepTop{
		rm: rm,
		ai: ai,
	}
}

// GetTopMiners returns n miners using the configured Reputation Module and
// Ask Index.
func (rt *RepTop) GetMiners(n int) ([]fpa.MinerProposal, error) {
	ms, err := rt.rm.GetTopMiners(n)
	if err != nil {
		return nil, fmt.Errorf("getting top %d miners from reputation module: %s", n, err)
	}
	aidx := rt.ai.Get()
	res := make([]fpa.MinerProposal, 0, len(ms))
	for _, m := range ms {
		sa, ok := aidx.Storage[m.Addr]
		if !ok {
			continue
		}
		res = append(res, fpa.MinerProposal{
			Addr:       sa.Miner,
			EpochPrice: sa.Price,
		})
	}
	return res, nil
}
