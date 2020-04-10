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
	GetCids() ([]cid.Cid, error)
}

// Config has general information about a Api instance.
type Config struct {
	ID               ffs.ApiID
	WalletAddr       string
	DefaultCidConfig ffs.DefaultCidConfig
}

// InstanceInfo has general information about a running Api instance.
type InstanceInfo struct {
	ID               ffs.ApiID
	DefaultCidConfig ffs.DefaultCidConfig
	Wallet           WalletInfo
	Pins             []cid.Cid
}

// WalletInfo contains information about the Wallet associated with
// the Api instance.
type WalletInfo struct {
	Address string
	Balance uint64
}

type GetLogsConfig struct {
	jid ffs.JobID
}

// GetLogsOption is a function that changes GetLogsConfig.
type GetLogsOption func(config *GetLogsConfig)

func WithJidFilter(jid ffs.JobID) GetLogsOption {
	return func(c *GetLogsConfig) {
		c.jid = jid
	}
}
