package api

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// PushConfigOption mutates a push configuration.
type PushConfigOption func(o *PushConfig) error

// PushConfig contains options for pushing a Cid configuration.
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

// WithCidConfig overrides the Api default Cid configuration.
func WithCidConfig(c ffs.CidConfig) PushConfigOption {
	return func(o *PushConfig) error {
		o.Config = c
		return nil
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushConfigOption {
	return func(o *PushConfig) error {
		o.OverrideConfig = override
		return nil
	}
}

// Validate validates a PushConfig.
func (pc PushConfig) Validate() error {
	if err := pc.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %s", err)
	}
	return nil
}
