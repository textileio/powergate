package ffs

import (
	"context"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car"
)

// WalletManager provides access to a Lotus wallet for a Lotus node.
type WalletManager interface {
	NewWallet(context.Context, string) (string, error)
	Balance(context.Context, string) (uint64, error)
}

// Scheduler enforces a CidConfig orchestrating Hot and Cold storages.
type Scheduler interface {
	PushConfig(PushConfigAction) (JobID, error)

	GetCidInfo(cid.Cid) (CidInfo, error)
	GetCidFromHot(context.Context, cid.Cid) (io.Reader, error)

	GetJob(JobID) (Job, error)
	Watch(InstanceID) <-chan Job
	Unwatch(<-chan Job)
}

// HotStorage is a fast datastorage layer for storing and retrieving raw
// data or Cids.
type HotStorage interface {
	Add(context.Context, io.Reader) (cid.Cid, error)
	Get(context.Context, cid.Cid) (io.Reader, error)
	Pin(context.Context, cid.Cid) (int, error)
	Put(context.Context, blocks.Block) error
}

// ColdStorage is a slow datastorage layer for storing Cids.
type ColdStorage interface {
	Store(context.Context, cid.Cid, string, FilConfig) (FilInfo, error)
	Retrieve(context.Context, cid.Cid, car.Store, string) (cid.Cid, error)
}

// MinerSelector returns miner addresses and ask storage information using a
// desired strategy.
type MinerSelector interface {
	GetMiners(int, MinerSelectorFilter) ([]MinerProposal, error)
}

// MinerSelectorFilter establishes filters that should be considered when
// returning miners.
type MinerSelectorFilter struct {
	// Blacklist contains miner names that should not be considered in
	// returned results. An empty list means no blacklisting.
	Blacklist []string
	// CountryCodes contains long-ISO country names that should be
	// considered in selected miners. An empty list means no filtering.
	CountryCodes []string
}

// MinerProposal contains a miners address and storage ask information
// to make a, most probably, successful deal.
type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}
