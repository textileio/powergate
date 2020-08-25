package api

import (
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

type ImportCheckError struct {
	cause string
}

func (ice *ImportCheckError) Error() string {
	return ice.cause
}

type ImportDeal struct {
	ProposalCid  *cid.Cid
	MinerAddress string
}

func (i *API) ImportStorage(payloadCid cid.Cid, pieceCid cid.Cid, deals []ImportDeal, opts ...ImportOption) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	cfg := importConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	scfg := ffs.StorageConfig{
		Repairable: false,
		Hot: ffs.HotConfig{
			Enabled: false,
		},
		Cold: i.cfg.DefaultStorageConfig.Cold,
	}

	filStorage := make([]ffs.FilStorage, len(deals))
	for i, d := range deals {
		filStorage[i] = ffs.FilStorage{
			PieceCid: pieceCid,
			Miner:    d.MinerAddress,
		}
		if d.ProposalCid != nil {
			filStorage[i].ProposalCid = *d.ProposalCid
		}
	}

	cinfo := ffs.CidInfo{
		JobID:   ffs.EmptyJobID,
		Cid:     payloadCid,
		Created: time.Now(),
		Hot:     ffs.HotInfo{Enabled: false},
		Cold: ffs.ColdInfo{
			Enabled: true,
			Filecoin: ffs.FilInfo{
				DataCid:   payloadCid,
				Proposals: filStorage,
			},
		},
	}

	// ToDo: use cfg.validate to validate import? And or fill missing info.
	if err := i.sched.ImportCidInfo(cinfo); err != nil {
		return fmt.Errorf("importing cid info in scheduler: %s", err)
	}

	if err := i.is.putStorageConfig(payloadCid, scfg); err != nil {
		return fmt.Errorf("saving new imported config: %s", err)
	}

	return nil
}
