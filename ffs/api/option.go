package api

import "github.com/textileio/powergate/ffs"

type AddCidOption func(o *AddCidConfig) error

type AddCidConfig struct {
	Config         ffs.CidConfig
	OverrideConfig bool
}

func newDefaultAddCidConfig(c ffs.CidConfig) AddCidConfig {
	return AddCidConfig{
		Config: c,
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
