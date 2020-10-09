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

// QueuedStorageJobs returns a list of queued storage jobs.
func (a *AdminService) QueuedStorageJobs(ctx context.Context, req *proto.QueuedStorageJobsRequest) (*proto.QueuedStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.QueuedStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.QueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (a *AdminService) ExecutingStorageJobs(ctx context.Context, req *proto.ExecutingStorageJobsRequest) (*proto.ExecutingStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.ExecutingStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.ExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *AdminService) LatestFinalStorageJobs(ctx context.Context, req *proto.LatestFinalStorageJobsRequest) (*proto.LatestFinalStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.LatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *AdminService) LatestSuccessfulStorageJobs(ctx context.Context, req *proto.LatestSuccessfulStorageJobsRequest) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := rpc.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.LatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (a *AdminService) StorageJobsSummary(ctx context.Context, req *proto.StorageJobsSummaryRequest) (*proto.StorageJobsSummaryResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	queuedJobs := a.s.QueuedStorageJobs(ffs.APIID(req.InstanceId), cids...)
	executingJobs := a.s.ExecutingStorageJobs(ffs.APIID(req.InstanceId), cids...)
	latestFinalJobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.InstanceId), cids...)
	latestSuccessfulJobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.InstanceId), cids...)

	protoQueuedJobs, err := rpc.ToProtoStorageJobs(queuedJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting queued jobs to protos: %v", err)
	}
	protoExecutingJobs, err := rpc.ToProtoStorageJobs(executingJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting executing jobs to protos: %v", err)
	}
	protoLatestFinalJobs, err := rpc.ToProtoStorageJobs(latestFinalJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest final jobs to protos: %v", err)
	}
	protoLatestSuccessfulJobs, err := rpc.ToProtoStorageJobs(latestSuccessfulJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest successful jobs to protos: %v", err)
	}

	return &proto.StorageJobsSummaryResponse{
		JobCounts: &rpc.JobCounts{
			Executing:        int32(len(executingJobs)),
			LatestFinal:      int32(len(latestFinalJobs)),
			LatestSuccessful: int32(len(latestSuccessfulJobs)),
			Queued:           int32(len(queuedJobs)),
		},
		ExecutingStorageJobs:        protoExecutingJobs,
		LatestFinalStorageJobs:      protoLatestFinalJobs,
		LatestSuccessfulStorageJobs: protoLatestSuccessfulJobs,
		QueuedStorageJobs:           protoQueuedJobs,
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
