package scheduler

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler/internal/astore"
)

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
	j := ffs.Job{
		ID:          jid,
		APIID:       iid,
		RetrievalID: rid,
		Status:      ffs.Queued,
	}

	ctx := context.WithValue(context.Background(), ffs.CtxKeyJid, jid)
	s.l.LogRetrieval(ctx, rid, "Scheduling new retrieval...")

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

	if err := s.sjs.Enqueue(j); err != nil {
		return ffs.EmptyJobID, fmt.Errorf("enqueuing job: %s", err)
	}

	select {
	case s.rd.evaluateQueue <- struct{}{}:
	default:
	}

	s.l.LogRetrieval(ctx, rid, "Retrieval scheduled successfully")
	return jid, nil
}
