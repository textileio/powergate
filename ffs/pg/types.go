package pg

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/fil-tools/ffs"
)

var (
	// ErrConfigNotFound returned when instance configuration doesn't exist
	// in ConfigStore.
	ErrConfigNotFound = errors.New("config not found")
	// ErrCidConfigNotFound returned when cid configuration doesn't exist in
	// ConfigStore.
	ErrCidConfigNotFound = errors.New("cid config not found")
	// ErrCidInfoNotFound returned if no storing state exists for a Cid.
	ErrCidInfoNotFound = errors.New("the cid doesn't have any saved state")
)

// ConfigStore is a repository for all state of a Powergate.
type ConfigStore interface {
	SaveConfig(c Config) error
	GetConfig() (*Config, error)

	GetCidConfig(cid.Cid) (*ffs.CidConfig, error)
	PushCidConfig(ffs.CidConfig) error

	SaveCidInfo(ffs.CidInfo) error
	GetCidInfo(cid.Cid) (ffs.CidInfo, error)
	Cids() ([]cid.Cid, error)
}

// Config has general information about a Powergate instance.
type Config struct {
	ID         ffs.InstanceID
	WalletAddr string
}

// InstanceInfo has general information about a running Powergate instance.
type InstanceInfo struct {
	ID     ffs.InstanceID
	Wallet WalletInfo
	Pins   []cid.Cid
}

// WalletInfo contains information about the Wallet associated with
// the Powergate instance.
type WalletInfo struct {
	Address string
	Balance uint64
}
