package api

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrConfigNotFound returned when instance configuration doesn't exist
	// in ConfigStore.
	ErrNotFound = errors.New("stored item not found")
)

// ConfigStore is a repository for all state of a Api.
type InstanceStore interface {
	PutConfig(c Config) error
	GetConfig() (Config, error)

	GetCidConfig(cid.Cid) (ffs.CidConfig, error)
	PutCidConfig(ffs.CidConfig) error
	GetCids() ([]cid.Cid, error)
}

// Config has general information about a Api instance.
type Config struct {
	ID               ffs.InstanceID
	WalletAddr       string
	DefaultCidConfig ffs.DefaultCidConfig
}

// InstanceInfo has general information about a running Api instance.
type InstanceInfo struct {
	ID               ffs.InstanceID
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
