package deals

import (
	"os"

	"github.com/ipfs/go-cid"
)

// StorageDealConfig contains information about a storage proposal for a miner
type StorageDealConfig struct {
	Miner      string
	EpochPrice uint64
}

type StoreResult struct {
	ProposalCid cid.Cid
	Config      StorageDealConfig
	Success     bool
}

// DealInfo contains information about a proposal storage deal
type DealInfo struct {
	ProposalCid cid.Cid
	StateID     uint64
	StateName   string
	Miner       string

	PieceCID cid.Cid
	Size     uint64

	PricePerEpoch uint64
	Duration      uint64

	DealID          uint64
	ActivationEpoch int64
}

type Config struct {
	ImportPath string
}

type Option func(*Config) error

func WithImportPath(path string) Option {
	return func(c *Config) error {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
		c.ImportPath = path
		return nil
	}
}
