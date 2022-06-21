package notifications

import (
	"io"

	"github.com/textileio/powergate/v2/ffs"
)

type JobUpdates interface {
	JobID() ffs.JobID
	FinalUpdates() bool

	Payload() (io.Reader, error)
	MatchNotificationEvent(event string) bool
	MatchNotificationAlert(alert *ffs.WebhookAlert) bool
}
