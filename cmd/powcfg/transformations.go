package main

import (
	"github.com/textileio/powergate/ffs"
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
