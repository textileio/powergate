package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateStorageProfile creates a new managed instance.
func (a *Service) CreateStorageProfile(ctx context.Context, req *proto.CreateStorageProfileRequest) (*proto.CreateStorageProfileResponse, error) {
	id, token, err := a.m.Create(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating instance: %v", err)
	}
	return &proto.CreateStorageProfileResponse{
		AuthEntry: &proto.AuthEntry{
			Id:    id.String(),
			Token: token,
		},
	}, nil
}

// ListStorageProfiles lists all managed instances.
func (a *Service) ListStorageProfiles(ctx context.Context, req *proto.ListStorageProfilesRequest) (*proto.ListStorageProfilesResponse, error) {
	lst, err := a.m.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing storage profiles: %v", err)
	}
	ins := make([]*proto.AuthEntry, len(lst))
	for i, v := range lst {
		ins[i] = &proto.AuthEntry{
			Id:    v.APIID.String(),
			Token: v.Token,
		}
	}
	return &proto.ListStorageProfilesResponse{
		AuthEntries: ins,
	}, nil
}
