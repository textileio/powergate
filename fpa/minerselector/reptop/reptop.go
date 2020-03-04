package reptop

import (
	"fmt"

	"github.com/textileio/fil-tools/fpa"
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

func (rt *RepTop) GetTopMiners(n int) ([]fpa.MinerProposal, error) {
	ms, err := rt.rm.GetTopMiners(n)
	if err != nil {
		return nil, fmt.Errorf("getting top N miners from reputation: %s", err)
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

var _ fpa.MinerSelector = (*RepTop)(nil)
