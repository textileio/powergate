package fixed

import (
	"fmt"

	"github.com/textileio/fil-tools/fpa/types"
)

type FixedMinerSelector struct {
	addr       string
	epochPrice uint64
}

func New(addr string, epochPrice uint64) *FixedMinerSelector {
	return &FixedMinerSelector{
		addr:       addr,
		epochPrice: epochPrice,
	}
}

func (fms *FixedMinerSelector) GetTopMiners(n int) ([]types.MinerProposal, error) {
	if n != 1 {
		return nil, fmt.Errorf("fixed miner selector can only return 1 result")
	}
	return []types.MinerProposal{
		{
			Addr:       fms.addr,
			EpochPrice: fms.epochPrice,
		},
	}, nil
}

var _ types.MinerSelector = (*FixedMinerSelector)(nil)
