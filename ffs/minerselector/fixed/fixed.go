package fixed

import (
	"fmt"

	"github.com/textileio/powergate/ffs"
)

// FixedMinerSelector is a ffs.MinerSelector implementation which
// always return a single miner address with an fixed epochPrice.
type FixedMinerSelector struct {
	addrs      []string
	epochPrice uint64
}

var _ ffs.MinerSelector = (*FixedMinerSelector)(nil)

// New returns a new FixedMinerSelector that always return addr as the miner address
// and epochPrice.
func New(addrs []string, epochPrice uint64) *FixedMinerSelector {
	return &FixedMinerSelector{
		addrs:      addrs,
		epochPrice: epochPrice,
	}
}

// GetTopMiners returns the single allowed miner in the selector.
func (fms *FixedMinerSelector) GetMiners(n int) ([]ffs.MinerProposal, error) {
	if n > len(fms.addrs) {
		return nil, fmt.Errorf("not enough fixed miners to provide, want %d, got %d", n, len(fms.addrs))
	}
	res := make([]ffs.MinerProposal, n)
	for i := 0; i < n; i++ {
		res[i] = ffs.MinerProposal{
			Addr:       fms.addrs[i%len(fms.addrs)],
			EpochPrice: fms.epochPrice,
		}
	}
	return res, nil
}
