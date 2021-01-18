package user

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	logger "github.com/ipfs/go-log/v2"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/buildinfo"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/api"
	"github.com/textileio/powergate/v2/ffs/manager"
	"github.com/textileio/powergate/v2/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	log = logger.Logger("user-service")
)

// Service implements the Powergate API.
type Service struct {
	userPb.UnimplementedUserServiceServer
	m   *manager.Manager
	w   wallet.Module
	hot ffs.HotStorage
}

// New creates a new powergate Service.
func New(m *manager.Manager, w wallet.Module, hot ffs.HotStorage) *Service {
	return &Service{
		m:   m,
		w:   w,
		hot: hot,
	}
}

// BuildInfo returns information about the powergate build.
func (s *Service) BuildInfo(ctx context.Context, req *userPb.BuildInfoRequest) (*userPb.BuildInfoResponse, error) {
	return &userPb.BuildInfoResponse{
		BuildDate:  buildinfo.BuildDate,
		GitBranch:  buildinfo.GitBranch,
		GitCommit:  buildinfo.GitCommit,
		GitState:   buildinfo.GitState,
		GitSummary: buildinfo.GitSummary,
		Version:    buildinfo.Version,
	}, nil
}

// UserIdentifier returns the API instance id.
func (s *Service) UserIdentifier(ctx context.Context, req *userPb.UserIdentifierRequest) (*userPb.UserIdentifierResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	id := i.ID()
	return &userPb.UserIdentifierResponse{Id: id.String()}, nil
}

func (s *Service) getInstanceByToken(ctx context.Context) (*api.API, error) {
	token := metautils.ExtractIncoming(ctx).Get("X-ffs-Token")
	if token == "" {
		return nil, status.Errorf(codes.PermissionDenied, "auth token can't be empty")
	}
	i, err := s.m.GetByAuthToken(token)
	if err != nil {
		code := codes.Internal
		if err == manager.ErrAuthTokenNotFound {
			code = codes.PermissionDenied
		}
		return nil, status.Errorf(code, "getting instance: %v", err)
	}
	return i, nil
}
