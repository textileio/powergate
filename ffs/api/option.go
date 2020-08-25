package api

import (
	"fmt"

	"github.com/textileio/powergate/ffs"
)

// PushStorageConfigOption mutates a push configuration.
type PushStorageConfigOption func(o *PushStorageConfigConfig) error

// PushStorageConfigConfig contains options for pushing a Cid configuration.
type PushStorageConfigConfig struct {
	Config         ffs.StorageConfig
	OverrideConfig bool
}

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c ffs.StorageConfig) PushStorageConfigOption {
	return func(o *PushStorageConfigConfig) error {
		o.Config = c
		return nil
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushStorageConfigOption {
	return func(o *PushStorageConfigConfig) error {
		o.OverrideConfig = override
		return nil
	}
}

// Validate validates a PushStorageConfigConfig.
func (pc PushStorageConfigConfig) Validate() error {
	if err := pc.Config.Validate(); err != nil {
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

type importConfig struct {
	validate bool
}

type ImportOption func(*importConfig)

func WithValidateImport(enabled bool) ImportOption {
	return func(c *importConfig) {
		c.validate = enabled
	}
}
