package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateUser creates a new managed instance.
func (a *Service) CreateUser(ctx context.Context, req *adminPb.CreateUserRequest) (*adminPb.CreateUserResponse, error) {
	auth, err := a.m.Create(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating instance: %v", err)
	}
	return &adminPb.CreateUserResponse{
		User: &adminPb.User{
			Id:    auth.APIID.String(),
			Token: auth.Token,
		},
	}, nil
}

// RegenerateAuth invalidates an existing token replacing it with a new one.
func (a *Service) RegenerateAuth(ctx context.Context, req *adminPb.RegenerateAuthRequest) (*adminPb.RegenerateAuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token can't be empty")
	}

	newToken, err := a.m.RegenerateAuthToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating instance: %v", err)
	}

	return &adminPb.RegenerateAuthResponse{
		NewToken: newToken,
	}, nil
}

// Users lists all managed instances.
func (a *Service) Users(ctx context.Context, req *adminPb.UsersRequest) (*adminPb.UsersResponse, error) {
	lst, err := a.m.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing users: %v", err)
	}
	ins := make([]*adminPb.User, len(lst))
	for i, v := range lst {
		ins[i] = &adminPb.User{
			Id:    v.APIID.String(),
			Token: v.Token,
		}
	}
	return &adminPb.UsersResponse{
		Users: ins,
	}, nil
}
