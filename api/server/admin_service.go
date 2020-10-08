package server

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/rpc"
	"github.com/textileio/powergate/ffs/scheduler"
	proto "github.com/textileio/powergate/proto/admin/v1"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminService implements the Admin API.
type AdminService struct {
	m *manager.Manager
	s *scheduler.Scheduler
}

// NewAdminService creates a new AdminService.
func NewAdminService(m *manager.Manager, s *scheduler.Scheduler) *AdminService {
	return &AdminService{
		m: m,
		s: s,
	}
}

// CreateInstance creates a new managed instance.
func (a *AdminService) CreateInstance(ctx context.Context, req *proto.CreateInstanceRequest) (*proto.CreateInstanceResponse, error) {
	id, token, err := a.m.Create(ctx)
	if err != nil {
		log.Errorf("creating instance: %s", err)
		return nil, status.Errorf(codes.Internal, "creating instance: %v", err)
	}
	return &proto.CreateInstanceResponse{
		Id:    id.String(),
		Token: token,
	}, nil
}

// ListInstances lists all managed instances.
func (a *AdminService) ListInstances(ctx context.Context, req *proto.ListInstancesRequest) (*proto.ListInstancesResponse, error) {
	lst, err := a.m.List()
	if err != nil {
		log.Errorf("listing instances: %s", err)
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

// GetQueuedStorageJobs returns a list of queued storage jobs.
func (a *AdminService) GetQueuedStorageJobs(ctx context.Context, req *proto.GetQueuedStorageJobsRequest) (*proto.GetQueuedStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.GetQueuedStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.GetQueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// GetExecutingStorageJobs returns a list of executing storage jobs.
func (a *AdminService) GetExecutingStorageJobs(ctx context.Context, req *proto.GetExecutingStorageJobsRequest) (*proto.GetExecutingStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.GetExecutingStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.GetExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// GetLatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *AdminService) GetLatestFinalStorageJobs(ctx context.Context, req *proto.GetLatestFinalStorageJobsRequest) (*proto.GetLatestFinalStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.GetLatestFinalStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.GetLatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// GetLatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *AdminService) GetLatestSuccessfulStorageJobs(ctx context.Context, req *proto.GetLatestSuccessfulStorageJobsRequest) (*proto.GetLatestSuccessfulStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.GetLatestSuccessfulStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.GetLatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

func fromProtoCids(cids []string) ([]cid.Cid, error) {
	var res []cid.Cid
	for _, cid := range cids {
		cid, err := util.CidFromString(cid)
		if err != nil {
			return nil, err
		}
		res = append(res, cid)
	}
	return res, nil
}
