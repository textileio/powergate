package fastapi

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/fil-tools/fpa"
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

// ConfigStore is a repository for all state of a FastAPI.
type ConfigStore interface {
	SaveConfig(c Config) error
	GetConfig() (*Config, error)

	GetCidConfig(cid.Cid) (*fpa.CidConfig, error)
	PushCidConfig(fpa.CidConfig) error

	SaveCidInfo(fpa.CidInfo) error
	GetCidInfo(cid.Cid) (fpa.CidInfo, error)
	Cids() ([]cid.Cid, error)
}

// Config has general information about a FastAPI instance.
type Config struct {
	ID         fpa.InstanceID
	WalletAddr string
}

// InstanceInfo has general information about a running FastAPI instance.
type InstanceInfo struct {
	ID     fpa.InstanceID
	Wallet WalletInfo
	Pins   []cid.Cid
}

// WalletInfo contains information about the Wallet associated with
// the FastAPI instance.
type WalletInfo struct {
	Address string
	Balance uint64
}
