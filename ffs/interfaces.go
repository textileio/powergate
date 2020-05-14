package ffs

import (
	"context"
	"errors"
	"io"
	"math/big"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car"
)

// WalletManager provides access to a Lotus wallet for a Lotus node.
type WalletManager interface {
	// NewAddress creates a new address.
	NewAddress(context.Context, string) (string, error)
	// Balance returns the current balance for an address.
	Balance(context.Context, string) (uint64, error)
	// SendFil sends fil from one address to another
	SendFil(ctx context.Context, from string, to string, amount *big.Int) error
}

var (
	// ErrHotStorageDisabled returned when trying to fetch a Cid when disabled on Hot Storage.
	// To retrieve the data, is necessary to call unfreeze by enabling the Enabled flag in
	// the Hot Storage for that Cid.
	ErrHotStorageDisabled = errors.New("cid disabled in hot storage")
)

// Scheduler enforces a CidConfig orchestrating Hot and Cold storages.
type Scheduler interface {
	// PushConfig push a new or modified configuration for a Cid. It returns
	// the JobID which tracks the current state of execution of that task.
	PushConfig(APIID, CidConfig) (JobID, error)

	// PushReplace push a new or modified configuration for a Cid, replacing
	// an existing one. The replaced Cid will be unstored from the Hot Storage.
	// Also it will be untracked (refer to Untrack() to understand implications)
	PushReplace(APIID, CidConfig, cid.Cid) (JobID, error)

	// GetCidInfo returns the current Cid storing state. This state may be different
	// from CidConfig which is the *desired* state.
	GetCidInfo(cid.Cid) (CidInfo, error)

	// GetCidFromHot returns an Reader with the Cid data. If the data isn't in the Hot
	// Storage, it errors with ErrHotStorageDisabled.
	GetCidFromHot(context.Context, cid.Cid) (io.Reader, error)

	// GetJob gets the a Job.
	GetJob(JobID) (Job, error)

	// WatchJobs is a blocking method that sends to a channel state updates
	// for all Jobs created by an Instance. The ctx should be canceled when
	// to stop receiving updates.
	WatchJobs(context.Context, chan<- Job, APIID) error

	// WatchLogs writes new log entries from Cid related executions.
	// This is a blocking operation that should be canceled by canceling the
	// provided context.
	WatchLogs(context.Context, chan<- LogEntry) error

	// GetLogs returns all registered logs for a Cid.
	GetLogs(context.Context, cid.Cid) ([]LogEntry, error)

	//Untrack marks a Cid to be untracked for any background processes such as
	// deal renewal, or repairing.
	Untrack(cid.Cid) error
}

// HotStorage is a fast storage layer for Cid data.
type HotStorage interface {
	// Add adds io.Reader data ephemerally (not pinned).
	Add(context.Context, io.Reader) (cid.Cid, error)

	// Remove removes a stored Cid.
	Remove(context.Context, cid.Cid) error

	// Get retrieves a stored Cid data.
	Get(context.Context, cid.Cid) (io.Reader, error)

	// Store stores a Cid. If the data wasn't previously Added,
	// depending on the implementation it may use internal mechanisms
	// for pulling the data, e.g: IPFS network
	Store(context.Context, cid.Cid) (int, error)

	// Replace replaces a stored Cid with a new one. It's mostly
	// thought for mutating data doing this efficiently.
	Replace(context.Context, cid.Cid, cid.Cid) (int, error)

	// Put adds a raw block.
	Put(context.Context, blocks.Block) error

	// IsStore returns true if the Cid is stored, or false
	// otherwise.
	IsStored(context.Context, cid.Cid) (bool, error)
}

// DealError contains information about a failed deal.
type DealError struct {
	ProposalCid cid.Cid
	Miner       string
	Message     string
}

// ColdStorage is slow/cheap storage for Cid data. It has
// native support for Filecoin storage.
type ColdStorage interface {
	// Store stores a Cid using the provided configuration and
	// account address. It returns a slice of deal errors happened
	// during execution.
	Store(context.Context, cid.Cid, FilConfig) (FilInfo, []DealError, error)

	// Retrieve retrieves the data using an account address,
	// and store it in a CAR store.
	Retrieve(context.Context, cid.Cid, car.Store, string) (cid.Cid, error)

	// EnsureRenewals executes renewal logic for a Cid under a particular
	// configuration. It returns a slice of deal errors happened during execution.
	EnsureRenewals(context.Context, cid.Cid, FilInfo, FilConfig) (FilInfo, []DealError, error)

	// IsFIlDealActive returns true if the proposal Cid is active on chain;
	// returns false otherwise.
	IsFilDealActive(context.Context, cid.Cid) (bool, error)
}

// MinerSelector returns miner addresses and ask storage information using a
// desired strategy.
type MinerSelector interface {
	// GetMiners returns a specified amount of miners that satisfy
	// provided filters.
	GetMiners(int, MinerSelectorFilter) ([]MinerProposal, error)
}

// MinerSelectorFilter establishes filters that should be considered when
// returning miners.
type MinerSelectorFilter struct {
	// ExcludedMiners contains miner addresses that should not be considered in
	// returned results. An empty list means no exclusions.
	ExcludedMiners []string
	// TrustedMiners contains miner addresses that will be prioritized
	// if are available in the query result. If the number of expected
	// results exceeeds the number of trusted miners, the remaining amount
	// of results will be returned still applying the rest of the filters
	// and the MinerSelector sorting logic.
	TrustedMiners []string
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
