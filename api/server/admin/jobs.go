package admin

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/api/server/powergate"
	"github.com/textileio/powergate/ffs"
	adminProto "github.com/textileio/powergate/proto/admin/v1"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueuedStorageJobs returns a list of queued storage jobs.
func (a *Service) QueuedStorageJobs(ctx context.Context, req *adminProto.QueuedStorageJobsRequest) (*adminProto.QueuedStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.QueuedStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := powergate.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminProto.QueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (a *Service) ExecutingStorageJobs(ctx context.Context, req *adminProto.ExecutingStorageJobsRequest) (*adminProto.ExecutingStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.ExecutingStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := powergate.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminProto.ExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *Service) LatestFinalStorageJobs(ctx context.Context, req *adminProto.LatestFinalStorageJobsRequest) (*adminProto.LatestFinalStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := powergate.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminProto.LatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *Service) LatestSuccessfulStorageJobs(ctx context.Context, req *adminProto.LatestSuccessfulStorageJobsRequest) (*adminProto.LatestSuccessfulStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.InstanceId), cids...)
	protoJobs, err := powergate.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminProto.LatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (a *Service) StorageJobsSummary(ctx context.Context, req *adminProto.StorageJobsSummaryRequest) (*adminProto.StorageJobsSummaryResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	queuedJobs := a.s.QueuedStorageJobs(ffs.APIID(req.InstanceId), cids...)
	executingJobs := a.s.ExecutingStorageJobs(ffs.APIID(req.InstanceId), cids...)
	latestFinalJobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.InstanceId), cids...)
	latestSuccessfulJobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.InstanceId), cids...)

	protoQueuedJobs, err := powergate.ToProtoStorageJobs(queuedJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting queued jobs to protos: %v", err)
	}
	protoExecutingJobs, err := powergate.ToProtoStorageJobs(executingJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting executing jobs to protos: %v", err)
	}
	protoLatestFinalJobs, err := powergate.ToProtoStorageJobs(latestFinalJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest final jobs to protos: %v", err)
	}
	protoLatestSuccessfulJobs, err := powergate.ToProtoStorageJobs(latestSuccessfulJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest successful jobs to protos: %v", err)
	}

	return &adminProto.StorageJobsSummaryResponse{
		JobCounts: &proto.JobCounts{
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
