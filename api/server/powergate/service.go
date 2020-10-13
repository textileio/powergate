package powergate

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/manager"
	walletModule "github.com/textileio/powergate/wallet/module"
)

var (
	// ErrEmptyAuthToken is returned when the provided auth-token is unknown.
	ErrEmptyAuthToken = errors.New("auth token can't be empty")
)

// Service implements the Powergate API.
type Service struct {
	m *manager.Manager
	w *walletModule.Module
}

// New creates a new powergate Service.
func New(m *manager.Manager, w *walletModule.Module) *Service {
	return &Service{
		m: m,
		w: w,
	}
}

func (s *Service) getInstanceByToken(ctx context.Context) (*api.API, error) {
	token := metautils.ExtractIncoming(ctx).Get("X-ffs-Token")
	if token == "" {
		return nil, ErrEmptyAuthToken
	}
	i, err := s.m.GetByAuthToken(token)
	if err != nil {
		return nil, err
	}
	return i, nil
}
