package admin

import (
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	dealsModule "github.com/textileio/powergate/v2/deals/module"
	"github.com/textileio/powergate/v2/ffs/manager"
	"github.com/textileio/powergate/v2/ffs/scheduler"
	askIndex "github.com/textileio/powergate/v2/index/ask/runner"
	minerIndex "github.com/textileio/powergate/v2/index/miner/lotusidx"
	"github.com/textileio/powergate/v2/wallet"
)

// Service implements the Admin API.
type Service struct {
	adminPb.UnimplementedAdminServiceServer
	m  *manager.Manager
	s  *scheduler.Scheduler
	wm wallet.Module
	dm *dealsModule.Module
	mi *minerIndex.Index
	ai *askIndex.Runner
}

// New creates a new AdminService.
func New(m *manager.Manager, s *scheduler.Scheduler, wm wallet.Module, dm *dealsModule.Module, mi *minerIndex.Index, ai *askIndex.Runner) *Service {
	return &Service{
		m:  m,
		s:  s,
		wm: wm,
		dm: dm,
		mi: mi,
		ai: ai,
	}
}
