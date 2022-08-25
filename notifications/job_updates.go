package notifications

import (
	"io"

	"github.com/textileio/powergate/v2/ffs"
)

// JobUpdates - provides interface for job notifications
type JobUpdates interface {
	// JobID - expose job id
	JobID() ffs.JobID
	// FinalUpdates - indicates if it's a final job update
	FinalUpdates() bool

	Notification
}

// Notification - common interface for notifications (events and alerts)
type Notification interface {
	// Payload - forms notification payload
	Payload() (io.Reader, error)
	// MatchEvent - matches notification by events
	MatchEvent(event string) bool
	// MatchAlert - matches notification by alerts
	MatchAlert(alert *ffs.WebhookAlert) bool
}
