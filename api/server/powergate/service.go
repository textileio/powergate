package powergate

import (
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/lotus"
)

// Service implements the Powergate API.
type Service struct {
	s             *scheduler.Scheduler
	clientBuilder lotus.ClientBuilder
}

// New creates a new powergate Service.
func New(s *scheduler.Scheduler, clientBuilder lotus.ClientBuilder) *Service {
	return &Service{
		s:             s,
		clientBuilder: clientBuilder,
	}
}
