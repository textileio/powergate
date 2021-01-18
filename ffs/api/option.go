package api

import (
	"fmt"

	"github.com/textileio/powergate/v2/ffs"
)

// PushStorageConfigOption mutates a push configuration.
type PushStorageConfigOption func(o *pushStorageConfigConfig) error

type pushStorageConfigConfig struct {
	config         ffs.StorageConfig
	overrideConfig bool
	dealIDs        []uint64
	noExec         bool
}

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c ffs.StorageConfig) PushStorageConfigOption {
	return func(o *pushStorageConfigConfig) error {
		o.config = c
		return nil
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushStorageConfigOption {
	return func(o *pushStorageConfigConfig) error {
		o.overrideConfig = override
		return nil
	}
}

// WithDealImport provides active deals to create/augment the storage info
// of the cid.
func WithDealImport(dealIDs []uint64) PushStorageConfigOption {
	return func(o *pushStorageConfigConfig) error {
		o.dealIDs = dealIDs
		return nil
	}
}

// WithNoExec avoids creating a Job for the new configuration.
func WithNoExec(noExec bool) PushStorageConfigOption {
	return func(o *pushStorageConfigConfig) error {
		o.noExec = noExec
		return nil
	}
}

// Validate validates a PushStorageConfigConfig.
func (pc pushStorageConfigConfig) Validate() error {
	if err := pc.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}

// NewAddressOption is a function that changes a NewAddressConfig.
type NewAddressOption func(config *NewAddressConfig)

// WithMakeDefault specifies if the new address should become the default.
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(c *NewAddressConfig) {
		c.makeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create.
func WithAddressType(addressType string) NewAddressOption {
	return func(c *NewAddressConfig) {
		c.addressType = addressType
	}
}

// GetLogsOption is a function that changes GetLogsConfig.
type GetLogsOption func(config *GetLogsConfig)

// WithJidFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJidFilter(jid ffs.JobID) GetLogsOption {
	return func(c *GetLogsConfig) {
		c.jid = jid
	}
}

// WithHistory indicates that prior log history should be sent
// to the channel before getting realtime logs.
func WithHistory(enabled bool) GetLogsOption {
	return func(c *GetLogsConfig) {
		c.history = enabled
	}
}

// RetrievalOption provides a retrieval configuration setup.
type RetrievalOption func(*retrievalConfig)

// WithRetrievalWalletAddress indicates which wallet address to use
// for doing the deal retrieval.
func WithRetrievalWalletAddress(addr string) RetrievalOption {
	return func(prc *retrievalConfig) {
		prc.walletAddress = addr
	}
}

// WithRetrievalMaxPrice indicates which is the maximum prices
// to pay for the retrieval.
func WithRetrievalMaxPrice(maxPrice uint64) RetrievalOption {
	return func(prc *retrievalConfig) {
		prc.maxPrice = maxPrice
	}
}
