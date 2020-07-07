package deals

import "os"

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
