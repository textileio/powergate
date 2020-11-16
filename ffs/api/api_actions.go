package api

import (
	"context"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// PushStorageConfig push a new configuration for the Cid in the hot and
// cold storage. If WithOverride opt isn't set it errors with ErrMustOverrideConfig.
func (i *API) PushStorageConfig(c cid.Cid, opts ...PushStorageConfigOption) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg := PushStorageConfigConfig{Config: i.cfg.DefaultStorageConfig}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return ffs.EmptyJobID, fmt.Errorf("config option: %s", err)
		}
	}
	if !cfg.OverrideConfig {
		_, err := i.is.getStorageConfigs(c)
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

	jid, err := i.sched.PushConfig(i.cfg.ID, c, cfg.Config)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling cid %s: %s", c, err)
	}
	if err := i.is.putStorageConfig(c, cfg.Config); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c, err)
	}
	return jid, nil
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (i *API) Remove(c cid.Cid) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfgs, err := i.is.getStorageConfigs(c)
	if err == ErrNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("getting cid config from store: %s", err)
	}
	if cfgs[c].Hot.Enabled || cfgs[c].Cold.Enabled {
		return ErrActiveInStorage
	}
	if err := i.sched.Untrack(c); err != nil {
		return fmt.Errorf("untracking from scheduler: %s", err)
	}
	if err := i.is.removeStorageConfig(c); err != nil {
		return fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return nil
}

// Replace pushes a StorageConfig for c2 equal to that of c1, and removes c1. This operation
// is more efficient than manually removing and adding in two separate operations.
// c1 and c2 must not be equal.
func (i *API) Replace(c1 cid.Cid, c2 cid.Cid) (ffs.JobID, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if c1.Equals(c2) {
		return ffs.EmptyJobID, fmt.Errorf("the old and new cid should be different")
	}

	cfgs, err := i.is.getStorageConfigs(c1)
	if err == ErrNotFound {
		return ffs.EmptyJobID, ErrReplacedCidNotFound
	}
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("getting replaced cid config: %s", err)
	}

	if err := i.ensureValidColdCfg(cfgs[c1].Cold); err != nil {
		return ffs.EmptyJobID, err
	}

	jid, err := i.sched.PushReplace(i.cfg.ID, c2, cfgs[c1], c1)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("scheduling replacement %s to %s: %s", c1, c2, err)
	}
	if err := i.is.putStorageConfig(c2, cfgs[c1]); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving new config for cid %s: %s", c2, err)
	}
	if err := i.is.removeStorageConfig(c1); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("deleting replaced cid config: %s", err)
	}
	return jid, nil
}

// Get returns an io.Reader for reading a stored Cid from hot storage.
func (i *API) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	if !c.Defined() {
		return nil, fmt.Errorf("cid is undefined")
	}
	cfgs, err := i.is.getStorageConfigs(c)
	if err != nil {
		return nil, fmt.Errorf("getting cid config: %s", err)
	}
	if !cfgs[c].Hot.Enabled {
		return nil, ErrHotStorageDisabled
	}
	r, err := i.sched.GetCidFromHot(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("getting from hot storage %s: %s", c, err)
	}
	return r, nil
}

// CancelJob cancels an executing Job. If no Job is executing
// with that JobID, it won't fail.
func (i *API) CancelJob(jid ffs.JobID) error {
	if err := i.sched.Cancel(jid); err != nil {
		return fmt.Errorf("canceling job %s: %s", jid, err)
	}
	return nil
}

func (i *API) ensureValidColdCfg(cfg ffs.ColdConfig) error {
	if cfg.Enabled && !i.isManagedAddress(cfg.Filecoin.Addr) {
		return fmt.Errorf("%v is not managed by ffs instance", cfg.Filecoin.Addr)
	}
	return nil
}
