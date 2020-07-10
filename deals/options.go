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
	IncludeFinal   bool
	Ascending      bool
}

// ListDealRecordsOption updates a ListDealRecordsConfig.
type ListDealRecordsOption func(*ListDealRecordsConfig)

// WithFromAddrs limits the results deals initiated from the provided wallet addresses.
// If WithDataCids is also provided, this is an AND operation.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
// If WithFromAddrs is also provided, this is an AND operation.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithIncludeFinal specifies whether or not to include final deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludeFinal(includeFinal bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.IncludeFinal = includeFinal
	}
}

// WithAscending specifies to sort the results in ascending order. Default is descending order.
// Records are sorted by timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}
