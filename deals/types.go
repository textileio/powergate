package deals

import (
	"github.com/ipfs/go-cid"
)

// StorageDealConfig contains information about a storage proposal for a miner.
type StorageDealConfig struct {
	Miner           string
	EpochPrice      uint64
	FastRetrieval   bool
	DealStartOffset int64
}

// StoreResult contains information about Executing deals.
type StoreResult struct {
	ProposalCid cid.Cid
	Config      StorageDealConfig
	Success     bool
	Message     string
}

// StorageDealInfo contains information about a proposed storage deal.
type StorageDealInfo struct {
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

// StorageDealRecord represents a storage deal log record.
type StorageDealRecord struct {
	RootCid           cid.Cid
	Addr              string
	DealInfo          StorageDealInfo
	Time              int64
	Pending           bool
	DataTransferStart int64
	DataTransferEnd   int64
	SealingStart      int64
	SealingEnd        int64
	ErrMsg            string
}

// RetrievalDealInfo contains information about a retrieval deal.
type RetrievalDealInfo struct {
	RootCid                 cid.Cid
	Size                    uint64
	MinPrice                uint64
	PaymentInterval         uint64
	PaymentIntervalIncrease uint64
	Miner                   string
	MinerPeerID             string
}

// RetrievalDealRecord represents a retrieval deal log record.
type RetrievalDealRecord struct {
	Addr              string
	DealInfo          RetrievalDealInfo
	Time              int64
	DataTransferStart int64
	DataTransferEnd   int64
	ErrMsg            string
}
