package fixed

import (
	"fmt"

	"github.com/textileio/powergate/ffs"
)

// FixedMinerSelector is a ffs.MinerSelector implementation which
// always return a single miner address with an fixed epochPrice.
type FixedMinerSelector struct {
	miners []Miner
}

type Miner struct {
	Addr       string
	Country    string
	EpochPrice uint64
}

var _ ffs.MinerSelector = (*FixedMinerSelector)(nil)

// New returns a new FixedMinerSelector that always return addr as the miner address
// and epochPrice.
func New(miners []Miner) *FixedMinerSelector {
	fixedMiners := make([]Miner, len(miners))
	copy(fixedMiners, miners)
	return &FixedMinerSelector{
		miners: fixedMiners,
	}
}

// GetTopMiners returns the single allowed miner in the selector.
func (fms *FixedMinerSelector) GetMiners(n int, f ffs.MinerSelectorFilter) ([]ffs.MinerProposal, error) {
	res := make([]ffs.MinerProposal, 0, n)
	for _, m := range fms.miners {
		skip := false
		for _, bAddr := range f.Blacklist {
			if bAddr == m.Addr {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if len(f.CountryCodes) != 0 {
			skip := true
			for _, c := range f.CountryCodes {
				if c == m.Country {
					skip = false
					break
				}
			}
			if skip {
				continue
			}
		}
		res = append(res, ffs.MinerProposal{
			Addr:       m.Addr,
			EpochPrice: m.EpochPrice,
		})

		if len(res) == n {
			break
		}
	}
	if len(res) != n {
		return nil, fmt.Errorf("not enough fixed miners to provide, want %d, got %d", n, len(res))
	}
	return res, nil
}
