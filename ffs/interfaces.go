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

// Scheduler creates and manages Job which executes Cid configurations
// in Hot and Cold layers, enables retrieval from those layers, and
// allows watching for Job state changes.
type Scheduler interface {
	EnqueueCid(AddAction) (JobID, error)
	GetCidFromHot(ctx context.Context, c cid.Cid) (io.Reader, error)

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
	GetMiners(n int) ([]MinerProposal, error)
}

// MinerProposal contains a miners address and storage ask information
// to make a, most probably, successful deal.
type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}
