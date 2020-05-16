package reptop

import (
	"fmt"

	"github.com/textileio/powergate/ffs"
	askRunner "github.com/textileio/powergate/index/ask/runner"
	"github.com/textileio/powergate/reputation"
)

// RepTop is a ffs.MinerSelector implementation that returns the top N
// miners from a Reputations Module and an Ask Index.
type RepTop struct {
	rm *reputation.Module
	ai *askRunner.Runner
}

var _ ffs.MinerSelector = (*RepTop)(nil)

// New returns a new RetTop instance that uses the specified Reputation Module
// to select miners and the AskIndex for their epoch prices.
func New(rm *reputation.Module, ai *askRunner.Runner) *RepTop {
	return &RepTop{
		rm: rm,
		ai: ai,
	}
}

// GetMiners returns n miners using the configured Reputation Module and
// Ask Index.
func (rt *RepTop) GetMiners(n int, f ffs.MinerSelectorFilter) ([]ffs.MinerProposal, error) {
	ms, err := rt.rm.QueryMiners(n, f.ExcludedMiners, f.CountryCodes, f.TrustedMiners)
	if err != nil {
		return nil, fmt.Errorf("getting top %d miners from reputation module: %s", n, err)
	}
	if len(ms) < n {
		return nil, fmt.Errorf("not enough miners that satisfy the constraints")
	}
	aidx := rt.ai.Get()
	res := make([]ffs.MinerProposal, 0, len(ms))
	for _, m := range ms {
		sa, ok := aidx.Storage[m.Addr]
		if !ok {
			continue
		}
		res = append(res, ffs.MinerProposal{
			Addr:       sa.Miner,
			EpochPrice: sa.Price,
		})
	}
	return res, nil
}
