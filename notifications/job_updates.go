package notifications

import (
	"io"

	"github.com/textileio/powergate/v2/ffs"
)

type JobUpdates interface {
	JobID() ffs.JobID
	FinalUpdates() bool

	Notification
}

type Notification interface {
	Payload() (io.Reader, error)
	MatchEvent(event string) bool
	MatchAlert(alert *ffs.WebhookAlert) bool
}
