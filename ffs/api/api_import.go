package api

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

//func (i *API) PushStorageConfig(c cid.Cid, opts ...PushStorageConfigOption) (ffs.JobID, error) {
func (i *API) PushImportConfig(payloadCid cid.Cid, pieceCid cid.Cid, deals []uint64, opts ...ImportOption) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	conf := importConfig{}
	for _, o := range opts {
		o(&conf)
	}

	// The StorageConfig of the import is the default storage config
	// but with a twist in HotConfig depending on opts.
	scfg := i.cfg.DefaultStorageConfig
	scfg.Hot.Enabled = false
	if conf.unfreeze != nil {
		scfg.Hot = ffs.HotConfig{
			Enabled:          true,
			AllowUnfreeze:    true,
			UnfreezeMaxPrice: conf.unfreeze.maxPrice,
		}
	}

	jid, err := i.sched.PushImport(i.cfg.ID, scfg)
	if err != nil {
		return fmt.Errorf("pushing import config: %s", err)
	}
	if err := i.is.putStorageConfig(payloadCid, scfg); err != nil {
		return fmt.Errorf("saving new imported config: %s", err)
	}

	return nil
}
