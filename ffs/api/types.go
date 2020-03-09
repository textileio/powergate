package api

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
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

// ConfigStore is a repository for all state of a Api.
type ConfigStore interface {
	SaveConfig(c Config) error
	GetConfig() (*Config, error)

	GetCidConfig(cid.Cid) (*ffs.AddAction, error)
	PushCidConfig(ffs.AddAction) error

	SaveCidInfo(ffs.CidInfo) error
	GetCidInfo(cid.Cid) (ffs.CidInfo, error)
	Cids() ([]cid.Cid, error)
}

// Config has general information about a Api instance.
type Config struct {
	ID            ffs.InstanceID
	WalletAddr    string
	DefaultConfig ffs.CidConfig
}

// InstanceInfo has general information about a running Api instance.
type InstanceInfo struct {
	ID     ffs.InstanceID
	Wallet WalletInfo
	Pins   []cid.Cid
}

// WalletInfo contains information about the Wallet associated with
// the Api instance.
type WalletInfo struct {
	Address string
	Balance uint64
}
