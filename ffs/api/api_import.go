package api

import (
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

type DealID uint64

func (i *API) ImportStorage(payloadCid cid.Cid, pieceCid cid.Cid, deals []DealID, opts ...ImportOption) error {
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

	// Generate StorageInfo as if Powergate saved this data.
	si, err := i.genStorageInfo(payloadCid, pieceCid, deals)
	if err != nil {
		return fmt.Errorf("generating storage info from imported deals: %s", err)
	}
	if err := i.sched.ImportStorageInfo(i.cfg.ID, si); err != nil {
		return fmt.Errorf("importing cid info in scheduler: %s", err)
	}
	if err := i.is.putStorageConfig(payloadCid, scfg); err != nil {
		return fmt.Errorf("saving new imported config: %s", err)
	}

	// We're done; we're in the same state as if Powergate
	// stored the data. If the user decided that wanted to
	// unfreeze the data (via opts), do it.
	if conf.unfreeze != nil {
		if err := i.scheduleImportJob(); err != nil {
			return fmt.Errorf("scheduling unfreeze job from import: %s", err)
		}
	}

	return nil
}

// Re-create FilStorage, as if they were created by Powergate
// storing the data itself.
func (i *API) genStorageInfo(payloadCid, pieceCid cid.Cid, deals []DealID) (ffs.StorageInfo, error) {
	fss := make([]ffs.FilStorage, len(deals))
	for j, dealID := range deals {
		fs, err := i.genFilStorage(dealID)
		if err != nil {
			return ffs.StorageInfo{}, fmt.Errorf("generating fil storage for dealid %d: %s", err)
		}
		if fs.PieceCid != pieceCid {
			return ffs.StorageInfo{}, fmt.Errorf("deal PieceCID doesn't match: %s != %s", fs.PieceCid, pieceCid)
		}
		fss[j] = fs
	}

	return ffs.StorageInfo{
		APIID:   i.cfg.ID,
		JobID:   ffs.EmptyJobID,
		Cid:     payloadCid,
		Created: time.Now(),
		Hot:     ffs.HotInfo{Enabled: false},
		Cold: ffs.ColdInfo{
			Enabled: true,
			Filecoin: ffs.FilInfo{
				DataCid:   payloadCid,
				Proposals: fss,
			},
		},
	}, nil
}

// genFilStorage generates FilStorage given a DealID.
func (i *API) genFilStorage(dealID DealID) (ffs.FilStorage, error) {
	panic("TODO")
	return ffs.FilStorage{
		DealID: uint64(dealID),
	}, nil

}
