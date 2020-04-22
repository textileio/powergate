package api

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound returned when instance configuration doesn't exist.
	ErrNotFound = errors.New("stored item not found")
)

// InstanceStore is a repository for all state of a Api.
type InstanceStore interface {
	PutConfig(c Config) error
	GetConfig() (Config, error)

	GetCidConfig(cid.Cid) (ffs.CidConfig, error)
	PutCidConfig(ffs.CidConfig) error
	RemoveCidConfig(cid.Cid) error
	GetCids() ([]cid.Cid, error)
}

// Config has general information about a Api instance.
type Config struct {
	ID            ffs.APIID
	Addrs         map[string]AddrInfo
	DefaultConfig ffs.DefaultConfig
}

// AddrInfo provides information about a wallet address
type AddrInfo struct {
	Name string
	Addr string
}

// InstanceInfo has general information about a running Api instance.
type InstanceInfo struct {
	ID            ffs.APIID
	DefaultConfig ffs.DefaultConfig
	Balances      []BalanceInfo
	Pins          []cid.Cid
}

// BalanceInfo contains the balance for the associated wallet address.
type BalanceInfo struct {
	AddrInfo
	Balance uint64
}

// NewAddressConfig contains options for creating a new wallet address
type NewAddressConfig struct {
	makeDefault bool
	addressType string
}

// NewAddressOption is a function that changes a NewAddressConfig
type NewAddressOption func(config *NewAddressConfig)

// WithMakeDefault specifies if the new address should become the default
func WithMakeDefault(makeDefault bool) NewAddressOption {
	return func(c *NewAddressConfig) {
		c.makeDefault = makeDefault
	}
}

// WithAddressType specifies the type of address to create
func WithAddressType(addressType string) NewAddressOption {
	return func(c *NewAddressConfig) {
		c.addressType = addressType
	}
}

// GetLogsConfig contains configuration for a stream-log
// of human-friendly messages for a Cid execution.
type GetLogsConfig struct {
	jid ffs.JobID
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
