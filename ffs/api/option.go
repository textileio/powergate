package api

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

type PushConfigOption func(o *PushConfig) error

type PushConfig struct {
	Config         ffs.CidConfig
	OverrideConfig bool
}

func newDefaultPushConfig(c cid.Cid, dc ffs.DefaultCidConfig) PushConfig {
	return PushConfig{
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

func WithCidConfig(c ffs.CidConfig) PushConfigOption {
	return func(o *PushConfig) error {
		o.Config = c
		return nil
	}
}

func WithOverride(override bool) PushConfigOption {
	return func(o *PushConfig) error {
		o.OverrideConfig = override
		return nil
	}
}

func (pc PushConfig) Validate() error {
	if err := pc.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}
