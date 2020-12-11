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
	// 1. Get current StorageInfo.
	si, err := s.cis.Get(iid, payloadCid)
	if err == cistore.ErrNotFound {
		si = ffs.StorageInfo{
			APIID:   iid,
			JobID:   ffs.EmptyJobID,
			Cid:     payloadCid,
			Created: time.Now(),
			Cold: ffs.ColdInfo{
				Filecoin: ffs.FilInfo{
					DataCid: payloadCid,
				},
			},
		}
	} else if err != nil {
		return fmt.Errorf("getting current storageinfo: %s", err)
	}

	// 2. Augment the retrieved/bootstraped StorageInfo with provided deals.
	if err := s.augmentStorageInfo(&si.Cold.Filecoin, dealIDs); err != nil {
		return fmt.Errorf("generating storage info from imported deals: %s", err)
	}
	if err := s.cis.Put(si); err != nil {
		return fmt.Errorf("importing cid information: %s", err)
	}

	return nil
}

func (s *Scheduler) augmentStorageInfo(fi *ffs.FilInfo, deals []uint64) error {
	// Deduplicate any already known DealID existing in StorageInfo with the
	// imported ones.
	var newDealIDs []uint64
	existingDealID := map[uint64]struct{}{}
	for _, p := range fi.Proposals {
		existingDealID[p.DealID] = struct{}{}
	}
	for _, did := range deals {
		if _, ok := existingDealID[did]; !ok {
			newDealIDs = append(newDealIDs, did)
		}
	}

	// If we have known deals, get the PieceCid. This will be used
	// for validating that the provided DealID are from deals also
	// for the same data. If that isn't the case, those DealIDs are
	// wrong to import.
	// If no known deals are present, we'll at least verify that all
	// DealIDs point to the same PieceCid, so at least are all
	// coherent. Unfortunately, there isn't a safe way to match
	// PieceCid for PayloadCid without the data.
	var pieceCid cid.Cid
	if len(fi.Proposals) > 0 {
		pieceCid = fi.Proposals[0].PieceCid
	}

	// Iterate new imported deals, and add the information
	// to StorageInfo as if Powergate made those deals.
	for _, dealID := range newDealIDs {
		fs, size, err := s.genFilStorage(dealID)
		if err != nil {
			return fmt.Errorf("generating fil storage for dealid %d: %s", dealID, err)
		}

		// If we couldn't know the right PieceCID from known deals
		// take the value from the first imported deals. Then use
		// this to also assert that other provided DealIDs are coherent
		// between each other.
		if pieceCid == cid.Undef {
			pieceCid = fs.PieceCid
		}

		// Validate that imported deals correspond all to the
		// same PieceCid. And if we already knew PieceCid from
		// previous deals Powergate did, also assert that's the case.
		if fs.PieceCid != pieceCid {
			return fmt.Errorf("invalid deal, PieceCID doesn't match: %s != %s", fs.PieceCid, pieceCid)
		}

		fi.Proposals = append(fi.Proposals, fs)

		// If our StorageInfo doesn't know about size, which is the case if we're boostrapping only
		// knowing DealIDs, then populate that data.
		if fi.Size == 0 {
			fi.Size = size
		}
	}

	return nil
}

// genFilStorage given a DealID it generates FilStorage and returns the piece size.
func (i *Scheduler) genFilStorage(dealID uint64) (ffs.FilStorage, uint64, error) {
	ctx, cls := context.WithTimeout(context.Background(), time.Second*10)
	defer cls()
	di, err := i.cs.GetDealInfo(ctx, dealID)
	if err != nil {
		return ffs.FilStorage{}, 0, fmt.Errorf("getting deal %d information: %s", dealID, err)
	}
	return ffs.FilStorage{
		DealID:     uint64(dealID),
		PieceCid:   di.Proposal.PieceCID,
		Renewed:    false,
		Duration:   int64(di.Proposal.EndEpoch) - int64(di.Proposal.StartEpoch) + 1,
		StartEpoch: uint64(di.Proposal.StartEpoch),
		Miner:      di.Proposal.Provider.String(),
		EpochPrice: di.Proposal.StoragePricePerEpoch.Uint64(),
	}, uint64(di.Proposal.PieceSize), nil
}
