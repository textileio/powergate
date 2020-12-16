package user

import (
	"context"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/api/server/util"
	su "github.com/textileio/powergate/api/server/util"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/manager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageJob calls API.GetStorageJob.
func (s *Service) StorageJob(ctx context.Context, req *userPb.StorageJobRequest) (*userPb.StorageJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	jid := ffs.JobID(req.JobId)
	job, err := i.GetStorageJob(jid)
	if err != nil {
		return nil, err
	}
	rpcJob, err := util.ToRPCJob(job)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "building job response: %v", err.Error())
	}
	return &userPb.StorageJobResponse{
		StorageJob: rpcJob,
	}, nil
}

// StorageConfigForJob returns the StorageConfig associated with the job id.
func (s *Service) StorageConfigForJob(ctx context.Context, req *userPb.StorageConfigForJobRequest) (*userPb.StorageConfigForJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		code := codes.Internal
		if err == manager.ErrAuthTokenNotFound || err == ErrEmptyAuthToken {
			code = codes.PermissionDenied
		}
		return nil, status.Errorf(code, "getting instance: %v", err)
	}
	sc, err := i.StorageConfigForJob(ffs.JobID(req.JobId))
	if err != nil {
		code := codes.Internal
		if err == api.ErrNotFound {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "getting storage config for job: %v", err)
	}
	res := toRPCStorageConfig(sc)
	return &userPb.StorageConfigForJobResponse{StorageConfig: res}, nil
}

// QueuedStorageJobs returns a list of queued storage jobs.
func (s *Service) QueuedStorageJobs(ctx context.Context, req *userPb.QueuedStorageJobsRequest) (*userPb.QueuedStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := su.FromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.QueuedStorageJobs(cids...)
	protoJobs, err := util.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &userPb.QueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (s *Service) ExecutingStorageJobs(ctx context.Context, req *userPb.ExecutingStorageJobsRequest) (*userPb.ExecutingStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := su.FromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.ExecutingStorageJobs(cids...)
	protoJobs, err := util.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &userPb.ExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (s *Service) LatestFinalStorageJobs(ctx context.Context, req *userPb.LatestFinalStorageJobsRequest) (*userPb.LatestFinalStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := su.FromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.LatestFinalStorageJobs(cids...)
	protoJobs, err := util.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &userPb.LatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (s *Service) LatestSuccessfulStorageJobs(ctx context.Context, req *userPb.LatestSuccessfulStorageJobsRequest) (*userPb.LatestSuccessfulStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := su.FromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.LatestSuccessfulStorageJobs(cids...)
	protoJobs, err := util.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &userPb.LatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (s *Service) StorageJobsSummary(ctx context.Context, req *userPb.StorageJobsSummaryRequest) (*userPb.StorageJobsSummaryResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := su.FromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	queuedJobs := i.QueuedStorageJobs(cids...)
	executingJobs := i.ExecutingStorageJobs(cids...)
	latestFinalJobs := i.LatestFinalStorageJobs(cids...)
	latestSuccessfulJobs := i.LatestSuccessfulStorageJobs(cids...)

	protoQueuedJobs, err := util.ToProtoStorageJobs(queuedJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting queued jobs to protos: %v", err)
	}
	protoExecutingJobs, err := util.ToProtoStorageJobs(executingJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting executing jobs to protos: %v", err)
	}
	protoLatestFinalJobs, err := util.ToProtoStorageJobs(latestFinalJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest final jobs to protos: %v", err)
	}
	protoLatestSuccessfulJobs, err := util.ToProtoStorageJobs(latestSuccessfulJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest successful jobs to protos: %v", err)
	}

	return &userPb.StorageJobsSummaryResponse{
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

// WatchStorageJobs calls API.WatchJobs.
func (s *Service) WatchStorageJobs(req *userPb.WatchStorageJobsRequest, srv userPb.UserService_WatchStorageJobsServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}

	jids := make([]ffs.JobID, len(req.JobIds))
	for i, jid := range req.JobIds {
		jids[i] = ffs.JobID(jid)
	}

	ch := make(chan ffs.StorageJob, 100)
	go func() {
		err = i.WatchJobs(srv.Context(), ch, jids...)
		close(ch)
	}()
	for job := range ch {
		rpcJob, err := util.ToRPCJob(job)
		if err != nil {
			return err
		}
		reply := &userPb.WatchStorageJobsResponse{
			StorageJob: rpcJob,
		}
		if err := srv.Send(reply); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// CancelStorageJob calls API.CancelJob.
func (s *Service) CancelStorageJob(ctx context.Context, req *userPb.CancelStorageJobRequest) (*userPb.CancelStorageJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	jid := ffs.JobID(req.JobId)
	if err := i.CancelJob(jid); err != nil {
		return nil, err
	}
	return &userPb.CancelStorageJobResponse{}, nil
}
