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

// DealInfo contains information about a proposed storage deal.
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

// DealRecord represents a storage deal log record.
type DealRecord struct {
	From     string
	DealInfo DealInfo
	Time     int64
	Pending  bool
}

// RetrievalInfo contains information about a retrieval deal.
type RetrievalInfo struct {
	PieceCID                cid.Cid
	Size                    uint64
	MinPrice                uint64
	PaymentInterval         uint64
	PaymentIntervalIncrease uint64
	Miner                   string
	MinerPeerID             string
}

// RetrievalRecord represents a retrieval deal log record.
type RetrievalRecord struct {
	From          string
	RetrievalInfo RetrievalInfo
	Time          int64
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
