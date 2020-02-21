package reptop

import (
	"fmt"

	"github.com/textileio/fil-tools/fpa/types"
	"github.com/textileio/fil-tools/index/ask"
	"github.com/textileio/fil-tools/reputation"
)

type RepTop struct {
	rm *reputation.Module
	ai *ask.AskIndex
}

func New(rm *reputation.Module, ai *ask.AskIndex) *RepTop {
	return &RepTop{
		rm: rm,
		ai: ai,
	}
}

func (rt *RepTop) GetTopMiners(n int) ([]types.MinerProposal, error) {
	ms, err := rt.rm.GetTopMiners(n)
	if err != nil {
		return nil, fmt.Errorf("getting top N miners from reputation: %s", err)
	}
	aidx := rt.ai.Get()
	res := make([]types.MinerProposal, 0, len(ms))
	for _, m := range ms {
		sa, ok := aidx.Storage[m.Addr]
		if !ok {
			continue
		}
		res = append(res, types.MinerProposal{
			Addr:       sa.Miner,
			EpochPrice: sa.Price,
		})
	}
	return res, nil
}

var _ types.MinerSelector = (*RepTop)(nil)
