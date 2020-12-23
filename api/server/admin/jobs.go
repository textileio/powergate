package admin

import (
	"context"

	"github.com/ipfs/go-cid"
	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/api/server/util"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueuedStorageJobs returns a list of queued storage jobs.
func (a *Service) QueuedStorageJobs(ctx context.Context, req *adminPb.QueuedStorageJobsRequest) (*adminPb.QueuedStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.QueuedStorageJobs(ffs.APIID(req.UserId), cids...)
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminPb.QueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (a *Service) ExecutingStorageJobs(ctx context.Context, req *adminPb.ExecutingStorageJobsRequest) (*adminPb.ExecutingStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.ExecutingStorageJobs(ffs.APIID(req.UserId), cids...)
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminPb.ExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *Service) LatestFinalStorageJobs(ctx context.Context, req *adminPb.LatestFinalStorageJobsRequest) (*adminPb.LatestFinalStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.UserId), cids...)
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminPb.LatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *Service) LatestSuccessfulStorageJobs(ctx context.Context, req *adminPb.LatestSuccessfulStorageJobsRequest) (*adminPb.LatestSuccessfulStorageJobsResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.UserId), cids...)
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &adminPb.LatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (a *Service) StorageJobsSummary(ctx context.Context, req *adminPb.StorageJobsSummaryRequest) (*adminPb.StorageJobsSummaryResponse, error) {
	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	queuedJobs := a.s.QueuedStorageJobs(ffs.APIID(req.UserId), cids...)
	executingJobs := a.s.ExecutingStorageJobs(ffs.APIID(req.UserId), cids...)
	latestFinalJobs := a.s.LatestFinalStorageJobs(ffs.APIID(req.UserId), cids...)
	latestSuccessfulJobs := a.s.LatestSuccessfulStorageJobs(ffs.APIID(req.UserId), cids...)

	protoQueuedJobs, err := su.ToProtoStorageJobs(queuedJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting queued jobs to protos: %v", err)
	}
	protoExecutingJobs, err := su.ToProtoStorageJobs(executingJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting executing jobs to protos: %v", err)
	}
	protoLatestFinalJobs, err := su.ToProtoStorageJobs(latestFinalJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest final jobs to protos: %v", err)
	}
	protoLatestSuccessfulJobs, err := su.ToProtoStorageJobs(latestSuccessfulJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest successful jobs to protos: %v", err)
	}

	return &adminPb.StorageJobsSummaryResponse{
		JobCounts: &userPb.JobCounts{
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
