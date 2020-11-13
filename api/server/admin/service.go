package admin

import (
	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/wallet"
)

// Service implements the Admin API.
type Service struct {
	adminPb.UnimplementedAdminServiceServer
	m  *manager.Manager
	s  *scheduler.Scheduler
	wm wallet.Module
}

// New creates a new AdminService.
func New(m *manager.Manager, s *scheduler.Scheduler, wm wallet.Module) *Service {
	return &Service{
		m:  m,
		s:  s,
		wm: wm,
	}
}
