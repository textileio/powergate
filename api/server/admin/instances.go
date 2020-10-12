package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateInstance creates a new managed instance.
func (a *Service) CreateInstance(ctx context.Context, req *proto.CreateInstanceRequest) (*proto.CreateInstanceResponse, error) {
	id, token, err := a.m.Create(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating instance: %v", err)
	}
	return &proto.CreateInstanceResponse{
		Id:    id.String(),
		Token: token,
	}, nil
}

// ListInstances lists all managed instances.
func (a *Service) ListInstances(ctx context.Context, req *proto.ListInstancesRequest) (*proto.ListInstancesResponse, error) {
	lst, err := a.m.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing instances: %v", err)
	}
	ins := make([]string, len(lst))
	for i, v := range lst {
		ins[i] = v.String()
	}
	return &proto.ListInstancesResponse{
		Instances: ins,
	}, nil
}
