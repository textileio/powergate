package notifications

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
)

type StorageJobUpdates struct {
	Job  ffs.StorageJob
	Info deals.StorageDealInfo
}

func (s StorageJobUpdates) JobID() ffs.JobID {
	return s.Job.ID
}

func (s StorageJobUpdates) FinalUpdates() bool {
	return false
}

func (s StorageJobUpdates) MatchAlert(alert *ffs.WebhookAlert) bool {
	return false
}

func (s StorageJobUpdates) MatchEvent(event string) bool {
	switch event {
	case AllEvents, AllStorageDealEvents:
		return true
	case AllCreatedEvents, StorageDealCreatedEvent:
		return s.Info.DealID != 0
	case AllCompletedEvents, StorageDealCompletedEvent:
		return s.Info.StateID == storagemarket.StorageDealActive

	case AllRetriedEvents, StorageDealRetriedEvent:
		// TODO:
		return false

	case AllFailedEvents, StorageDealFailedEvent:
		return s.Job.Status == ffs.Failed ||
			s.Info.StateID == storagemarket.StorageDealFailing ||
			s.Info.StateID == storagemarket.StorageDealError

	case AllCanceledEvents, StorageDealCanceledEvent:
		return s.Job.Status == ffs.Canceled

	case StorageDealExpiredEvent:
		return s.Info.StateID == storagemarket.StorageDealExpired

	case StorageDealSlashedEvent:
		return s.Info.StateID == storagemarket.StorageDealSlashed

	default:
		return false
	}
}

type storageJobNotification struct {
	Cid         string    `json:"cid"`
	JobID       ffs.JobID `json:"jobId"`
	JobStatus   string    `json:"jobStatus"`
	Miner       string    `json:"miner"`
	Price       uint64    `json:"price"`
	ProposalCid string    `json:"proposalCid"`
	DealID      uint64    `json:"dealId,omitempty"`
	DealStatus  string    `json:"dealStatus"`
	ErrCause    string    `json:"error,omitempty"`
	Message     string    `json:"message,omitempty"`
}

func (s StorageJobUpdates) Payload() (io.Reader, error) {
	obj := &storageJobNotification{
		Cid:         s.Job.Cid.String(),
		JobID:       s.Job.ID,
		JobStatus:   ffs.JobStatusStr[s.Job.Status],
		Miner:       s.Info.Miner,
		Price:       s.Info.PricePerEpoch,
		ProposalCid: s.Info.ProposalCid.String(),
		DealID:      s.Info.DealID,
		DealStatus:  s.Info.StateName,
		ErrCause:    s.Job.ErrCause,
		Message:     s.Info.Message,
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}
