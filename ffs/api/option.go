package api

import "github.com/textileio/powergate/ffs"

type AddCidOption func(o *AddCidConfig) error

type AddCidConfig struct {
	Config ffs.CidConfig
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
