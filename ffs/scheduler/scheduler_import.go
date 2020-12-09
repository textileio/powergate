package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/cistore"
)

func (s *Scheduler) ImportDeals(iid ffs.APIID, payloadCid cid.Cid, dealIDs []uint64) error {
	_, err := s.cis.Get(iid, payloadCid)
	if err != nil && err != cistore.ErrNotFound {
		return fmt.Errorf("checking if cid info already exists: %s", err)
	}
	if err != cistore.ErrNotFound {
		return fmt.Errorf("there is cid information for the provided cid")
	}

	// 1. Get current StorageInfo.
	si, err := s.cis.Get(iid, payloadCid)
	if err == cistore.ErrNotFound {
		si = ffs.StorageInfo{
			APIID:   iid,
			JobID:   ffs.EmptyJobID,
			Cid:     payloadCid,
			Created: time.Now(),
		}
	}
	if err != nil {
		return fmt.Errorf("getting current storageinfo: %s", err)
	}

	// 2. Augment the retrieved/bootstraped StorageInfo with provided deals.
	if err := s.augmentStorageInfo(&si, dealIDs); err != nil {
		return fmt.Errorf("generating storage info from imported deals: %s", err)
	}
	if err := s.cis.Put(si); err != nil {
		return fmt.Errorf("importing cid information: %s", err)
	}

	return nil
}

func (s *Scheduler) augmentStorageInfo(si *ffs.StorageInfo, deals []uint64) error {
	fss := make([]ffs.FilStorage, len(deals))
	for j, dealID := range deals {
		fs, err := s.genFilStorage(dealID)
		if err != nil {
			return ffs.StorageInfo{}, fmt.Errorf("generating fil storage for dealid %d: %s", dealID, err)
		}
		if fs.PieceCid != pieceCid {
			return ffs.StorageInfo{}, fmt.Errorf("invalid deal, PieceCID doesn't match: %s != %s", fs.PieceCid, pieceCid)
		}
		fss[j] = fs
	}

	return ffs.StorageInfo{
		APIID:   iid,
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
func (i *Scheduler) genFilStorage(dealID uint64) (ffs.FilStorage, error) {
	ctx, cls := context.WithTimeout(context.Background(), time.Second*10)
	defer cls()
	di, err := i.cs.GetDealInfo(ctx, dealID)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("getting deal %d information :%s", dealID, err)
	}
	return ffs.FilStorage{
		DealID:     uint64(dealID),
		PieceCid:   di.Proposal.PieceCID,
		Renewed:    false,
		Duration:   int64(di.Proposal.EndEpoch) - int64(di.Proposal.StartEpoch) + 1,
		StartEpoch: uint64(di.State.SectorStartEpoch),
		Miner:      di.Proposal.Provider.String(),
		EpochPrice: di.Proposal.StoragePricePerEpoch.Uint64(),
	}, nil
}
