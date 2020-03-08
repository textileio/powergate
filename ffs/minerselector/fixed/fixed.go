package fixed

import (
	"fmt"

	"github.com/textileio/fil-tools/ffs"
)

// FixedMinerSelector is a ffs.MinerSelector implementation which
// always return a single miner address with an fixed epochPrice.
type FixedMinerSelector struct {
	addr       string
	epochPrice uint64
}

var _ ffs.MinerSelector = (*FixedMinerSelector)(nil)

// New returns a new FixedMinerSelector that always return addr as the miner address
// and epochPrice.
func New(addr string, epochPrice uint64) *FixedMinerSelector {
	return &FixedMinerSelector{
		addr:       addr,
		epochPrice: epochPrice,
	}
}

// GetTopMiners returns the single allowed miner in the selector.
func (fms *FixedMinerSelector) GetMiners(n int) ([]ffs.MinerProposal, error) {
	if n != 1 {
		return nil, fmt.Errorf("fixed miner selector can only return 1 result")
	}
	return []ffs.MinerProposal{{
		Addr:       fms.addr,
		EpochPrice: fms.epochPrice,
	}}, nil
}
