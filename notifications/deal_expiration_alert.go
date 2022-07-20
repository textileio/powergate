package notifications

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/util"
)

// DealExpirationAlert - uses for alert notification when deal close to expiration and goes below threshold
type DealExpirationAlert struct {
	JobID        ffs.JobID
	DealID       uint64
	PieceCid     cid.Cid
	Miner        string
	ExpiryEpoch  uint64
	CurrentEpoch uint64
}

type dealExpirationAlertNotification struct {
	AlertType           string    `json:"alertType"`
	JobID               ffs.JobID `json:"jobId"`
	DealID              uint64    `json:"dealId"`
	PieceCid            string    `json:"pieceCid"`
	Miner               string    `json:"miner"`
	EpochTillExpiration uint64    `json:"epochTillExpiration"`
	Error               string    `json:"error"`
}

// Payload - forms notification payload
func (d DealExpirationAlert) Payload() (io.Reader, error) {
	var epochTillExpiration uint64
	msg := "deal already expired"

	if d.ExpiryEpoch > d.CurrentEpoch {
		msg = "deal close to expiration"
		epochTillExpiration = d.ExpiryEpoch - d.CurrentEpoch
	}

	obj := &dealExpirationAlertNotification{
		AlertType:           DealExpirationCheck,
		JobID:               d.JobID,
		DealID:              d.DealID,
		PieceCid:            d.PieceCid.String(),
		Miner:               d.Miner,
		EpochTillExpiration: epochTillExpiration,
		Error:               msg,
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data), nil
}

// MatchEvent - checks for matching with events
func (d DealExpirationAlert) MatchEvent(event string) bool {
	return false
}

// MatchAlert - checks for matching with alerts
func (d DealExpirationAlert) MatchAlert(alert *ffs.WebhookAlert) bool {
	if alert == nil {
		return false
	}

	if alert.Type != DealExpirationCheck {
		return false
	}

	threshold, err := time.ParseDuration(alert.Threshold)
	if err != nil {
		log.Errorf("failed to parse deal expiration check threshold: %s", err)
		return false
	}

	if d.ExpiryEpoch <= d.CurrentEpoch {
		// already expired
		return true
	}

	epochTillExpiration := d.ExpiryEpoch - d.CurrentEpoch
	epochThreshold := uint64(threshold.Seconds() / util.EpochDurationSeconds)

	return epochTillExpiration < epochThreshold
}
