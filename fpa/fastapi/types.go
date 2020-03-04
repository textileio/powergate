package fastapi

import (
	"errors"

	"github.com/ipfs/go-cid"
	"github.com/textileio/fil-tools/fpa"
)

type Config struct {
	ID         fpa.InstanceID
	WalletAddr string
}

type InstanceInfo struct {
	ID     fpa.InstanceID
	Wallet WalletInfo
	Pins   []cid.Cid
}

type WalletInfo struct {
	Address string
	Balance uint64
}

var (
	ErrConfigNotFound    = errors.New("config not found")
	ErrCidConfigNotFound = errors.New("cid config not found")
)

type ConfigStore interface {
	SaveConfig(c Config) error
	GetConfig() (*Config, error)

	GetCidConfig(cid.Cid) (*fpa.CidConfig, error)
	PushCidConfig(fpa.CidConfig) error

	SaveCidInfo(fpa.CidInfo) error
	GetCidInfo(cid.Cid) (fpa.CidInfo, bool, error)
	Cids() ([]cid.Cid, error)
}
