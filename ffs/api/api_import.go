package api

import (
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

// ImportDeal contains information of an imported deal.
type ImportDeal struct {
	ProposalCid  *cid.Cid
	MinerAddress string
}

// ImportStorage imports deals existing in the Filecoin network. The StorageConfig
// attached to this Cid will be the default one with HotStorage disabled.
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

	sinfo := ffs.StorageInfo{
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

	if err := i.sched.ImportCidInfo(sinfo); err != nil {
		return fmt.Errorf("importing cid info in scheduler: %s", err)
	}

	if err := i.is.putStorageConfig(payloadCid, scfg); err != nil {
		return fmt.Errorf("saving new imported config: %s", err)
	}

	return nil
}
