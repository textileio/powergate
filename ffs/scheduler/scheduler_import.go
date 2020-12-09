package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/cistore"
)

func (s *Scheduler) PushImport(iid ffs.APIID, payloadCid, pieceCid cid.Cid, dealIDs []uint64, cfg ffs.StorageConfig) (ffs.JobID, error) {
	_, err := s.cis.Get(iid, payloadCid)
	if err != nil && err != cistore.ErrNotFound {
		return ffs.EmptyJobID, fmt.Errorf("checking if cid info already exists: %s", err)
	}
	if err != cistore.ErrNotFound {
		return ffs.EmptyJobID, fmt.Errorf("there is cid information for the provided cid")
	}

	// Generate StorageInfo as if Powergate saved this data.
	si, err := s.genStorageInfo(iid, payloadCid, pieceCid, dealIDs)
	if err != nil {
		return ffs.EmptyJobID, fmt.Errorf("generating storage info from imported deals: %s", err)
	}
	if err := s.cis.Put(si); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("importing cid information: %s", err)
	}

	return s.push(iid, payloadCid, cfg, cid.Undef)
}

// Re-create FilStorage, as if they were created by Powergate
// storing the data itself.
func (s *Scheduler) genStorageInfo(iid ffs.APIID, payloadCid, pieceCid cid.Cid, deals []uint64) (ffs.StorageInfo, error) {
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
