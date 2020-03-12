package api

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

type AddCidOption func(o *AddCidConfig) error

type AddCidConfig struct {
	Config         ffs.CidConfig
	OverrideConfig bool
}

func newDefaultAddCidConfig(c cid.Cid, dc ffs.DefaultCidConfig) AddCidConfig {
	return AddCidConfig{
		Config: newDefaultCidConfig(c, dc),
	}
}

func newDefaultCidConfig(c cid.Cid, dc ffs.DefaultCidConfig) ffs.CidConfig {
	return ffs.CidConfig{
		Cid:  c,
		Hot:  dc.Hot,
		Cold: dc.Cold,
	}
}

func WithCidConfig(c ffs.CidConfig) AddCidOption {
	return func(o *AddCidConfig) error {
		o.Config = c
		return nil
	}
}

func WithOverride(override bool) AddCidOption {
	return func(o *AddCidConfig) error {
		o.OverrideConfig = override
		return nil
	}
}

func (acc AddCidConfig) Validate() error {
	if err := acc.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}
