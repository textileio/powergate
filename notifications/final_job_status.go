package notifications

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/textileio/powergate/v2/ffs"
)

// FinalJobStatus - uses for notification about final job status
type FinalJobStatus struct {
	JobId      ffs.JobID
	JobStatus  ffs.JobStatus
	JobError   error
	DealErrors []ffs.DealError
}

// JobID - expose job id
func (f FinalJobStatus) JobID() ffs.JobID {
	return f.JobId
}

// FinalUpdates - checks if it's final job update
func (f FinalJobStatus) FinalUpdates() bool {
	return true
}

// MatchAlert - matches for alerts
func (f FinalJobStatus) MatchAlert(alert *ffs.WebhookAlert) bool {
	return false
}

// MatchEvent - matches for events
func (f FinalJobStatus) MatchEvent(event string) bool {
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

// Payload - forms notification payload for final job update
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
