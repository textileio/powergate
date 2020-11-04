package powergate

import (
	"context"

	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/manager"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageJob calls API.GetStorageJob.
func (s *Service) StorageJob(ctx context.Context, req *proto.StorageJobRequest) (*proto.StorageJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	jid := ffs.JobID(req.JobId)
	job, err := i.GetStorageJob(jid)
	if err != nil {
		return nil, err
	}
	rpcJob, err := toRPCJob(job)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "building job response: %v", err.Error())
	}
	return &proto.StorageJobResponse{
		Job: rpcJob,
	}, nil
}

// StorageConfigForJob returns the StorageConfig associated with the job id.
func (s *Service) StorageConfigForJob(ctx context.Context, req *proto.StorageConfigForJobRequest) (*proto.StorageConfigForJobResponse, error) {
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
	res := ToRPCStorageConfig(sc)
	return &proto.StorageConfigForJobResponse{StorageConfig: res}, nil
}

// QueuedStorageJobs returns a list of queued storage jobs.
func (s *Service) QueuedStorageJobs(ctx context.Context, req *proto.QueuedStorageJobsRequest) (*proto.QueuedStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.QueuedStorageJobs(cids...)
	protoJobs, err := ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.QueuedStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (s *Service) ExecutingStorageJobs(ctx context.Context, req *proto.ExecutingStorageJobsRequest) (*proto.ExecutingStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.ExecutingStorageJobs(cids...)
	protoJobs, err := ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.ExecutingStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (s *Service) LatestFinalStorageJobs(ctx context.Context, req *proto.LatestFinalStorageJobsRequest) (*proto.LatestFinalStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.LatestFinalStorageJobs(cids...)
	protoJobs, err := ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.LatestFinalStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (s *Service) LatestSuccessfulStorageJobs(ctx context.Context, req *proto.LatestSuccessfulStorageJobsRequest) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}
	jobs := i.LatestSuccessfulStorageJobs(cids...)
	protoJobs, err := ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	return &proto.LatestSuccessfulStorageJobsResponse{
		StorageJobs: protoJobs,
	}, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (s *Service) StorageJobsSummary(ctx context.Context, req *proto.StorageJobsSummaryRequest) (*proto.StorageJobsSummaryResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	queuedJobs := i.QueuedStorageJobs(cids...)
	executingJobs := i.ExecutingStorageJobs(cids...)
	latestFinalJobs := i.LatestFinalStorageJobs(cids...)
	latestSuccessfulJobs := i.LatestSuccessfulStorageJobs(cids...)

	protoQueuedJobs, err := ToProtoStorageJobs(queuedJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting queued jobs to protos: %v", err)
	}
	protoExecutingJobs, err := ToProtoStorageJobs(executingJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting executing jobs to protos: %v", err)
	}
	protoLatestFinalJobs, err := ToProtoStorageJobs(latestFinalJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest final jobs to protos: %v", err)
	}
	protoLatestSuccessfulJobs, err := ToProtoStorageJobs(latestSuccessfulJobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting latest successful jobs to protos: %v", err)
	}

	return &proto.StorageJobsSummaryResponse{
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

// WatchStorageJobs calls API.WatchJobs.
func (s *Service) WatchStorageJobs(req *proto.WatchStorageJobsRequest, srv proto.PowergateService_WatchStorageJobsServer) error {
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
		rpcJob, err := toRPCJob(job)
		if err != nil {
			return err
		}
		reply := &proto.WatchStorageJobsResponse{
			Job: rpcJob,
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
func (s *Service) CancelStorageJob(ctx context.Context, req *proto.CancelStorageJobRequest) (*proto.CancelStorageJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	jid := ffs.JobID(req.JobId)
	if err := i.CancelJob(jid); err != nil {
		return nil, err
	}
	return &proto.CancelStorageJobResponse{}, nil
}
