package admin

import (
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/scheduler"
)

// Service implements the Admin API.
type Service struct {
	m *manager.Manager
	s *scheduler.Scheduler
}

// New creates a new AdminService.
func New(m *manager.Manager, s *scheduler.Scheduler) *Service {
	return &Service{
		m: m,
		s: s,
	}
}
