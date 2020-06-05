package deals

import (
	"os"

	"github.com/ipfs/go-cid"
)

// StorageDealConfig contains information about a storage proposal for a miner.
type StorageDealConfig struct {
	Miner      string
	EpochPrice uint64
}

// StoreResult contains information about Executing deals.
type StoreResult struct {
	ProposalCid cid.Cid
	Config      StorageDealConfig
	Success     bool
	Message     string
}

// DealInfo contains information about a proposal storage deal.
type DealInfo struct {
	ProposalCid cid.Cid
	StateID     uint64
	StateName   string
	Miner       string

	PieceCID cid.Cid
	Size     uint64

	PricePerEpoch uint64
	StartEpoch    uint64
	Duration      uint64

	DealID          uint64
	ActivationEpoch int64
	Message         string
}

// Config contains configuration for storing deals.
type Config struct {
	ImportPath string
}

// Option sets values on a Config.
type Option func(*Config) error

// WithImportPath indicates the import path that will be used
// to store data to later be imported to Lotus.
func WithImportPath(path string) Option {
	return func(c *Config) error {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
		c.ImportPath = path
		return nil
	}
}
