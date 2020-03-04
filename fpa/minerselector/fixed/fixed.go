package fixed

import (
	"fmt"

	"github.com/textileio/fil-tools/fpa"
)

// FixedMinerSelectr is a fpa.MinerSelector implementation which
// always return a single miner address with an fixed epochPrice.
type FixedMinerSelector struct {
	addr       string
	epochPrice uint64
}

var _ fpa.MinerSelector = (*FixedMinerSelector)(nil)

// New returns a new FixedMinerSelector that always return addr as the miner address
// and epochPrice.
func New(addr string, epochPrice uint64) *FixedMinerSelector {
	return &FixedMinerSelector{
		addr:       addr,
		epochPrice: epochPrice,
	}
}

// GetTopMiners returns the single allowed miner in the selector.
func (fms *FixedMinerSelector) GetTopMiners(n int) ([]fpa.MinerProposal, error) {
	if n != 1 {
		return nil, fmt.Errorf("fixed miner selector can only return 1 result")
	}
	return []fpa.MinerProposal{{
		Addr:       fms.addr,
		EpochPrice: fms.epochPrice,
	}}, nil
}
