package ffs

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
)

// WalletManager provides access to a Lotus wallet for a Lotus node.
type WalletManager interface {
	NewWallet(ctx context.Context, typ string) (string, error)
	Balance(ctx context.Context, addr string) (uint64, error)
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
	Pin(context.Context, cid.Cid, HotConfig) (HotInfo, error)
}

// ColdStorage is a slow datastorage layer for storing Cids.
type ColdStorage interface {
	Store(ctx context.Context, c cid.Cid, waddr string, conf ColdConfig) (ColdInfo, error)
}

// MinerSelector returns miner addresses and ask storage information using a
// desired strategy.
type MinerSelector interface {
	GetMiners(n int, f MinerSelectorFilter) ([]MinerProposal, error)
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
