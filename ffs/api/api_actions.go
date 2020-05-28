package api

import (
	"context"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
)

// PushConfig push a new configuration for the Cid in the Hot and
// Cold layer. If WithOverride opt isn't set it errors with ErrMustOverrideConfig
func (i *API) PushConfig(c cid.Cid, opts ...PushConfigOption) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg := newDefaultPushConfig(c, i.cfg.DefaultConfig)
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return ffs.EmptyJobID, fmt.Errorf("config option: %s", err)
		}
	}
	if !cfg.OverrideConfig {
		_, err := i.is.getCidConfig(c)
		if err == nil {
			return ffs.EmptyJobID, ErrMustOverrideConfig
		}
		if err != ErrNotFound {
			return ffs.EmptyJobID, fmt.Errorf("getting cid config: %s", err)
		}
	}
	if err := cfg.Config.Validate(); err != nil {
		return ffs.EmptyJobID, err
	}
	if err := i.ensureValidColdCfg(cfg.Config.Cold); err != nil {
		return ffs.EmptyJobID, err
	}

	jid, err := i.sched.PushConfig(i.cfg.ID, cfg.Config)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.is.putCidConfig(cfg.Config); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (i *API) Remove(c cid.Cid) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg, err := i.is.getCidConfig(c)
	if err == ErrNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("getting cid config from store: %s", err)
	}
	if cfg.Hot.Enabled || cfg.Cold.Enabled {
		return ErrActiveInStorage
	}
	if err := i.sched.Untrack(c); err != nil {
		return fmt.Errorf("untracking from scheduler: %s", err)
	}
	if err := i.is.removeCidConfig(c); err != nil {
		return fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return nil
}

// Replace pushes a CidConfig of c2 equal to c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
// c1 and c2 must not be equal.
func (i *API) Replace(c1 cid.Cid, c2 cid.Cid) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if c1.Equals(c2) {
		return ffs.EmptyJobID, fmt.Errorf("the old and new cid should be different")
	}

	cfg, err := i.is.getCidConfig(c1)
	if err == ErrNotFound {
		return ffs.EmptyJobID, ErrReplacedCidNotFound
	}
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("getting replaced cid config: %s", err)
	}
	cfg.Cid = c2

	if err := i.ensureValidColdCfg(cfg.Cold); err != nil {
		return ffs.EmptyJobID, err
	}

	jid, err := i.sched.PushReplace(i.cfg.ID, cfg, c1)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling replacement %s to %s: %s", c1, c2, err)
	}
	if err := i.is.putCidConfig(cfg); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c2, err)
	}
	if err := i.is.removeCidConfig(c1); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return jid, nil
}

// Get returns an io.Reader for reading a stored Cid from the Hot Storage.
func (i *API) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	if !c.Defined() {
		return nil, fmt.Errorf("cid is undefined")
	}
	conf, err := i.is.getCidConfig(c)
	if err != nil {
		return nil, fmt.Errorf("getting cid config: %s", err)
	}
	if !conf.Hot.Enabled {
		return nil, ErrHotStorageDisabled
	}
	r, err := i.sched.GetCidFromHot(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting from hot layer %s: %s", c, err)
	}
	return r, nil
}

// Show returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns ErrNotFound.
func (i *API) Show(c cid.Cid) (ffs.CidInfo, error) {
	inf, err := i.sched.GetCidInfo(c)
	if err == scheduler.ErrNotFound {
		return inf, ErrNotFound
	}
	if err != nil {
		return inf, fmt.Errorf("getting cid information: %s", err)
	}
	return inf, nil
}

func (i *API) ensureValidColdCfg(cfg ffs.ColdConfig) error {
	if cfg.Enabled && !i.isManagedAddress(cfg.Filecoin.Addr) {
		return fmt.Errorf("%v is not managed by ffs instance", cfg.Filecoin.Addr)
	}
	return nil
}
