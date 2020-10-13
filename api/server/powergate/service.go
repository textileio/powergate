package powergate

import (
	"github.com/textileio/powergate/ffs/scheduler"
	walletModule "github.com/textileio/powergate/wallet/module"
)

// Service implements the Powergate API.
type Service struct {
	s *scheduler.Scheduler
	w *walletModule.Module
}

// New creates a new powergate Service.
func New(s *scheduler.Scheduler, w *walletModule.Module) *Service {
	return &Service{
		s: s,
		w: w,
	}
}
