package notifications

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/textileio/powergate/v2/ffs"
)

type FinalJobStatus struct {
	JobId      ffs.JobID
	JobStatus  ffs.JobStatus
	JobError   error
	DealErrors []ffs.DealError
}

func (f FinalJobStatus) JobID() ffs.JobID {
	return f.JobId
}

func (f FinalJobStatus) FinalUpdates() bool {
	return true
}

func (f FinalJobStatus) MatchNotificationAlert(alert *ffs.WebhookAlert) bool {
	return false
}

func (f FinalJobStatus) MatchNotificationEvent(event string) bool {
	switch event {
	case AllEvents, AllStorageDealEvents:
		return true

	case AllCompletedEvents, StorageDealCompletedEvent:
		return f.JobStatus == ffs.Success

	case AllFailedEvents, StorageDealFailedEvent:
		return f.JobStatus == ffs.Failed

	case AllCanceledEvents, StorageDealCanceledEvent:
		return f.JobStatus == ffs.Canceled

	default:
		return false
	}
}

type finalStorageJobNotification struct {
	JobId      ffs.JobID   `json:"jobId"`
	JobStatus  string      `json:"jobStatus"`
	JobError   string      `json:"jobError,omitempty"`
	DealErrors []dealError `json:"dealErrors,omitempty"`
}

type dealError struct {
	ProposalCid string `json:"proposalCid"`
	Miner       string `json:"miner"`
	Message     string `json:"error"`
}

func (f FinalJobStatus) Payload() (io.Reader, error) {
	var errMessage string
	if f.JobError != nil {
		errMessage = f.JobError.Error()
	}

	var dealErrors []dealError
	for _, deal := range f.DealErrors {
		dealErrors = append(dealErrors, dealError{
			ProposalCid: deal.ProposalCid.String(),
			Miner:       deal.Miner,
			Message:     deal.Message,
		})
	}

	obj := &finalStorageJobNotification{
		JobId:      f.JobId,
		JobStatus:  ffs.JobStatusStr[f.JobStatus],
		JobError:   errMessage,
		DealErrors: dealErrors,
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}
