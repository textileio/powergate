package ffs

import (
	"context"
	"io"
	"math/big"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"
)

// WalletManager provides access to a Lotus wallet for a Lotus node.
type WalletManager interface {
	// NewAddress creates a new address.
	NewAddress(context.Context, string) (string, error)
	// Balance returns the current balance for an address.
	Balance(context.Context, string) (uint64, error)
	// SendFil sends fil from one address to another.
	SendFil(context.Context, string, string, *big.Int) error
}

// ListDealRecordsConfig specifies the options for DealsManager.List.
type ListDealRecordsConfig struct {
	FromAddrs      []string
	DataCids       []string
	IncludePending bool
	OnlyPending    bool
	Ascending      bool
}

// ListDealRecordsOption updates a ListDealRecordsConfig.
type ListDealRecordsOption func(*ListDealRecordsConfig)

// WithFromAddrs limits the results deals initated from the provided wallet addresses.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithOnlyPending specifies whether or not to only include pending deals in the results.
// Ignored for ListRetrievalDealRecords.
func WithOnlyPending(onlyPending bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.OnlyPending = onlyPending
	}
}

// WithAscending specifies to sort the results in ascending order.
// Default is descending order.
// If pending, records are sorted by timestamp, otherwise records
// are sorted by activation epoch then timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}

// DealsManager specifies the required deals api.
type DealsManager interface {
	ListStorageDealRecords(...ListDealRecordsOption) ([]deals.StorageDealRecord, error)
	ListRetrievalDealRecords(...ListDealRecordsOption) ([]deals.RetrievalDealRecord, error)
}

// PaychManager provides access to payment channels.
type PaychManager interface {
	// List lists all payment channels involving the specified addresses.
	List(ctx context.Context, addrs ...string) ([]PaychInfo, error)
	// Create creates a new payment channel.
	Create(ctx context.Context, from string, to string, amount uint64) (PaychInfo, cid.Cid, error)
	// Redeem redeems a payment channel.
	Redeem(ctx context.Context, ch string) error
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

func (de DealError) Error() string {
	return de.Message
}

// ColdStorage is slow/cheap storage for Cid data. It has
// native support for Filecoin storage.
type ColdStorage interface {
	// Store stores a Cid using the provided configuration and
	// account address. It returns a slice of accepted proposed deals,
	// a slice of rejected proposal deals, and the size of the data.
	Store(context.Context, cid.Cid, FilConfig) ([]cid.Cid, []DealError, uint64, error)

	// WaitForDeal blocks the provided Deal Proposal reach a
	// final state. If the deal finishes successfully it returns a FilStorage
	// result. If the deal finished with error, it returns a ffs.DealError
	// error result, so it should be considered in error handling.
	WaitForDeal(context.Context, cid.Cid, cid.Cid) (FilStorage, error)

	// Fetch fetches the cid data in the underlying storage.
	Fetch(context.Context, cid.Cid, string) error

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
	// MaxPrice is the max ask price to consider when selecting miner deals
	MaxPrice uint64
}

// MinerProposal contains a miners address and storage ask information
// to make a, most probably, successful deal.
type MinerProposal struct {
	Addr       string
	EpochPrice uint64
}
