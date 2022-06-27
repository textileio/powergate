package notifications

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/textileio/powergate/v2/ffs"
)

// RetrievalJobUpdates - uses for retrieval job events
type RetrievalJobUpdates struct {
	Job  ffs.RetrievalJob
	Info ffs.RetrievalInfo
}

func (r RetrievalJobUpdates) JobID() ffs.JobID {
	return r.Job.ID
}

func (r RetrievalJobUpdates) FinalUpdates() bool {
	return false
}

type retrievalJobNotification struct {
	JobID       ffs.JobID       `json:"jobId"`
	JobStatus   string          `json:"jobStatus"`
	RetrievalID ffs.RetrievalID `json:"retrievalId"`
	DataCid     string          `json:"dataCid"`
	TotalPaid   uint64          `json:"totalPaid"`
	Miner       string          `json:"miner"`
	Size        int64           `json:"size"`
	CreatedAt   string          `json:"createdAt"`
	ErrCause    string          `json:"error,omitempty"`
}

func (r RetrievalJobUpdates) Payload() (io.Reader, error) {
	obj := &retrievalJobNotification{
		JobID:       r.Job.ID,
		JobStatus:   ffs.JobStatusStr[r.Job.Status],
		RetrievalID: r.Job.RetrievalID,
		DataCid:     r.Info.DataCid.String(),
		TotalPaid:   r.Info.TotalPaid,
		Miner:       r.Info.MinerAddr,
		Size:        r.Info.Size,
		CreatedAt:   r.Info.CreatedAt.Format(time.RFC3339),
		ErrCause:    r.Job.ErrCause,
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

func (r RetrievalJobUpdates) MatchEvent(event string) bool {
	switch event {
	case AllEvents, AllDataRetrievalEvents:
		return true
	case AllCreatedEvents, DataRetrievalCreatedEvent:
		return r.Job.Status == ffs.Executing
	case AllCompletedEvents, DataRetrievalCompletedEvent:
		return r.Job.Status == ffs.Success

	case AllRetriedEvents, DataRetrievalRetriedEvent:
		// TODO:
		return false

	case AllFailedEvents, DataRetrievalFailedEvent:
		return r.Job.Status == ffs.Failed

	case AllCanceledEvents, DataRetrievalCanceledEvent:
		return r.Job.Status == ffs.Canceled

	default:
		return false
	}
}

func (r RetrievalJobUpdates) MatchAlert(alert *ffs.WebhookAlert) bool {
	return false
}
