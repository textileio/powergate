package main

import (
	"github.com/textileio/powergate/v2/ffs"
)

// bumpIpfsAddTimeout is a transformation that updates Hot.Ipfs.AddTimeout
// if the current value is less than a minimum desired value.
func bumpIpfsAddTimeout(minValue int) Transform {
	return func(cfg *ffs.StorageConfig) (bool, error) {
		if cfg.Hot.Ipfs.AddTimeout >= minValue {
			return false, nil
		}

		cfg.Hot.Ipfs.AddTimeout = minValue

		return true, nil
	}
}

var _ = unsetFilecoinUnlimitedPrice

// unsetFilecoinUnlimitedPrice is a transformation that sets
// Cold.Filecoin.MaxPrice to maxPrice if the current value is
// 0 (no limit).
func unsetFilecoinUnlimitedPrice(maxPrice uint64) Transform {
	return func(cfg *ffs.StorageConfig) (bool, error) {
		if cfg.Cold.Filecoin.MaxPrice > 0 {
			return false, nil
		}

		cfg.Cold.Filecoin.MaxPrice = maxPrice

		return true, nil
	}
}
