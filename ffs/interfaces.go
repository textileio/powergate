package ffs

import (
	"context"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/deals"
)

// WalletManager provides access to a Lotus wallet for a Lotus node.
type WalletManager interface {
	// MasterAddr returns the master address.
	// Will return address.Undef is Powergate was started with no master address.
	MasterAddr() address.Address
	// NewAddress creates a new address.
	NewAddress(context.Context, string) (string, error)
	// Balance returns the current balance for an address.
	Balance(context.Context, string) (*big.Int, error)
	// SendFil sends fil from one address to another.
	SendFil(context.Context, string, string, *big.Int) (cid.Cid, error)
	// Sign signs a message using an address.
	Sign(context.Context, string, []byte) ([]byte, error)
	// Verify verifies if a message was signed with an address.
	Verify(context.Context, string, []byte, []byte) (bool, error)
}

// DealRecordsManager provides access to deal records.
type DealRecordsManager interface {
	ListStorageDealRecords(opts ...deals.DealRecordsOption) ([]deals.StorageDealRecord, error)
	ListRetrievalDealRecords(opts ...deals.DealRecordsOption) ([]deals.RetrievalDealRecord, error)
}

// HotStorage is a fast storage layer for Cid data.
type HotStorage interface {
	// Stage adds io.Reader and stage-pins it.
	Stage(context.Context, APIID, io.Reader) (cid.Cid, error)

	// StageCid pulls Cid data and stage-pin it.
	StageCid(context.Context, APIID, cid.Cid) error

	// Unpin unpins a Cid.
	Unpin(context.Context, APIID, cid.Cid) error

	// Get retrieves a stored Cid data.
	Get(context.Context, cid.Cid) (io.Reader, error)

	// Pin pins a Cid. If the data wasn't previously Added,
	// depending on the implementation it may use internal mechanisms
	// for pulling the data, e.g: IPFS network
	Pin(context.Context, APIID, cid.Cid) (int, error)

	// Replace replaces a stored Cid with a new one. It's mostly
	// thought for mutating data doing this efficiently.
	Replace(context.Context, APIID, cid.Cid, cid.Cid) (int, error)

	// IsPinned returns true if the Cid is pinned, or false
	// otherwise.
	IsPinned(context.Context, APIID, cid.Cid) (bool, error)

	// GCStaged unpins Cids that are stage-pinned that aren't
	// contained in a exclude list, and were pinned before a time.
	GCStaged(context.Context, []cid.Cid, time.Time) ([]cid.Cid, error)

	// PinnedCids returns pinned cids information.
	PinnedCids(context.Context) ([]PinnedCid, error)
}

// DealError contains information about a failed deal.
type DealError struct {
	ProposalCid cid.Cid
	Miner       string
	Message     string
}

// Error returns an stringified message of the
// underlying error cause.
func (de DealError) Error() string {
	return de.Message
}

// FetchInfo describes information about a Filecoin retrieval.
type FetchInfo struct {
	RetrievedMiner string
	FundsSpent     uint64
}

var (
	// ErrOnChainDealNotFound is returned when a deal doesn't exist,
	// or might have been slashed.
	ErrOnChainDealNotFound = errors.New("on-chain deal not found, may not exist or have been slashed")
)

// ColdStorage is slow/cheap storage for Cid data. It has
// native support for Filecoin storage.
type ColdStorage interface {
	// Store stores a Cid using the provided configuration and
	// account address. It returns a slice of accepted proposed deals,
	// a slice of rejected proposal deals, and the size of the data.
	Store(context.Context, cid.Cid, FilConfig) ([]cid.Cid, []DealError, abi.PaddedPieceSize, error)

	// WaitForDeal blocks the provided Deal Proposal reach a
	// final state. If the deal finishes successfully it returns a FilStorage
	// result. If the deal finished with error, it returns a ffs.DealError
	// error result, so it should be considered in error handling.
	WaitForDeal(context.Context, cid.Cid, cid.Cid, time.Duration, chan deals.StorageDealInfo) (FilStorage, error)

	// Fetch fetches the cid data in the underlying storage.
	Fetch(context.Context, cid.Cid, *cid.Cid, string, []string, uint64, string) (FetchInfo, error)

	// EnsureRenewals executes renewal logic for a Cid under a particular
	// configuration. It returns a slice of deal errors happened during execution.
	EnsureRenewals(context.Context, cid.Cid, FilInfo, FilConfig, time.Duration, chan deals.StorageDealInfo) (FilInfo, []DealError, error)

	// GetDealInfo returns information about an on-chain deal.
	// If a deal ID doesn't exist, the deal isn't active anymore, or was
	// slashed, then ErrOnChainDealNotFound is returned.
	GetDealInfo(context.Context, uint64) (api.MarketDeal, error)

	// GetCurrentEpoch returns information about current epoch on-chain.
	GetCurrentEpoch(context.Context) (uint64, error)
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
	// MaxPrice is the max ask price to consider when selecting miner deals
	MaxPrice uint64
	// PieceSize is the piece size of the data.
	PieceSize uint64
	// VerifiedDeal indicates it should take verified storage prices.
	VerifiedDeal bool
}

// MinerProposal contains a miners address and storage ask information
// to make a, most probably, successful deal.
type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}

// PinnedCid provides information about a pinned Cid.
type PinnedCid struct {
	Cid    cid.Cid
	APIIDs []APIIDPinnedCid
}

// APIIDPinnedCid has information about a Cid pinned by a user.
type APIIDPinnedCid struct {
	ID        APIID
	Staged    bool
	CreatedAt int64
}
