package scheduler

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/astore"
	"github.com/textileio/powergate/ffs/scheduler/internal/ristore"
)

// StartRetrieval schedules a new RetrievalJob to execute a Filecoin retrieval.
func (s *Scheduler) StartRetrieval(iid ffs.APIID, rid ffs.RetrievalID, pyCid, piCid cid.Cid, sel string, miners []string, walletAddr string, maxPrice uint64) (ffs.JobID, error) {
	if iid == ffs.EmptyInstanceID {
		return ffs.EmptyJobID, fmt.Errorf("empty API ID")
	}
	if rid == ffs.EmptyRetrievalID {
		return ffs.EmptyJobID, fmt.Errorf("empty retrieval ID")
	}
	if pyCid == cid.Undef {
		return ffs.EmptyJobID, fmt.Errorf("payload cid is undefined")
	}
	if piCid == cid.Undef {
		return ffs.EmptyJobID, fmt.Errorf("piece cid is undefined")
	}
	if len(miners) == 0 {
		return ffs.EmptyJobID, fmt.Errorf("miner list can't be empty")
	}
	if walletAddr == "" {
		return ffs.EmptyJobID, fmt.Errorf("wallet address can't be empty")
	}

	jid := ffs.NewJobID()
	j := ffs.RetrievalJob{
		ID:          jid,
		APIID:       iid,
		RetrievalID: rid,
		Status:      ffs.Queued,
	}

	ctx := context.WithValue(context.Background(), ffs.CtxKeyJid, jid)
	ctx = context.WithValue(ctx, ffs.CtxRetrievalID, rid)
	s.l.Log(ctx, "Scheduling new retrieval...")

	ra := astore.RetrievalAction{
		APIID:         iid,
		RetrievalID:   rid,
		PayloadCid:    pyCid,
		PieceCid:      piCid,
		Selector:      sel,
		Miners:        miners,
		WalletAddress: walletAddr,
		MaxPrice:      maxPrice,
	}
	if err := s.as.PutRetrievalAction(j.ID, ra); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("saving retrieval action for job: %s", err)
	}

	if err := s.rjs.Enqueue(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("enqueuing job: %s", err)
	}

	select {
	case s.rd.evaluateQueue <- struct{}{}:
	default:
	}

	s.l.Log(ctx, "Retrieval scheduled successfully")
	return jid, nil
}

// GetRetrievalInfo returns the information about an executed retrieval.
func (s *Scheduler) GetRetrievalInfo(rid ffs.RetrievalID) (ffs.RetrievalInfo, error) {
	info, err := s.ris.Get(rid)
	if err == ristore.ErrNotFound {
		return ffs.RetrievalInfo{}, ErrNotFound
	}
	if err != nil {
		return ffs.RetrievalInfo{}, fmt.Errorf("getting retrieval info from store: %s", err)
	}
	return info, nil
}

func (s *Scheduler) executeRetrieval(ctx context.Context, a astore.RetrievalAction, j ffs.RetrievalJob) (ffs.RetrievalInfo, error) {
	// ToDo: WIP.
	_ = j
	fi, err := s.cs.Fetch(ctx, a.PayloadCid, &a.PieceCid, a.WalletAddress, a.Miners, a.MaxPrice, a.Selector)
	if err != nil {
		return ffs.RetrievalInfo{}, fmt.Errorf("fetching from cold storage: %s", err)
	}
	// ToDo: use graphsync to get the underlying DataCid to be pinned
	dataCid := cid.Undef

	size, err := s.hs.Pin(ctx, a.APIID, dataCid)
	if err != nil {
		return ffs.RetrievalInfo{}, fmt.Errorf("pinning data cid: %s", err)
	}

	ri := ffs.RetrievalInfo{
		ID:        a.RetrievalID,
		DataCid:   dataCid,
		TotalPaid: fi.FundsSpent,
		MinerAddr: fi.RetrievedMiner,
		Size:      int64(size),
	}

	return ri, nil
}
