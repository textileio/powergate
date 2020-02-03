package deals

import (
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"os"
)

// DealConfig contains information about a proposal for a particular miner
type DealConfig struct {
	Miner      string
	EpochPrice types.BigInt
}

// DealInfo contains information about a proposal storage deal
type DealInfo struct {
	ProposalCid cid.Cid
	StateID     uint64
	StateName   string
	Miner       string

	PieceRef []byte
	Size     uint64

	PricePerEpoch types.BigInt
	Duration      uint64
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
