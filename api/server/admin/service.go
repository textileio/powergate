package admin

import (
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"github.com/textileio/powergate/v2/deals/module"
	"github.com/textileio/powergate/v2/ffs/manager"
	"github.com/textileio/powergate/v2/ffs/scheduler"
	"github.com/textileio/powergate/v2/wallet"
)

// Service implements the Admin API.
type Service struct {
	adminPb.UnimplementedAdminServiceServer
	m  *manager.Manager
	s  *scheduler.Scheduler
	wm wallet.Module
	dm *module.Module
}

// New creates a new AdminService.
func New(m *manager.Manager, s *scheduler.Scheduler, wm wallet.Module, dm *module.Module) *Service {
	return &Service{
		m:  m,
		s:  s,
		wm: wm,
		dm: dm,
	}
}
