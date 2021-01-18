package user

import (
	"context"

	"github.com/ipfs/go-cid"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/api/server/util"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/util"
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
	rpcJob, err := su.ToRPCJob(job)
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
		return nil, err
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

// ListStorageJobs lists StorageJobs according to the provided request parameters.
func (s *Service) ListStorageJobs(ctx context.Context, req *userPb.ListStorageJobsRequest) (*userPb.ListStorageJobsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	var selector api.ListStorageJobsSelect
	switch req.Selector {
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_ALL:
		selector = api.All
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_EXECUTING:
		selector = api.Executing
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_FINAL:
		selector = api.Final
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_QUEUED:
		selector = api.Queued
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_UNSPECIFIED:
		selector = api.All
	}
	conf := api.ListStorageJobsConfig{
		Limit:         req.Limit,
		Ascending:     req.Ascending,
		NextPageToken: req.NextPageToken,
		Select:        selector,
	}
	if req.CidFilter != "" {
		c, err := cid.Decode(req.CidFilter)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "parsing cid filter: %v", err)
		}
		conf.CidFilter = c
	}
	jobs, more, next, err := i.ListStorageJobs(conf)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing storage jobs: %v", err)
	}
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	res := &userPb.ListStorageJobsResponse{
		StorageJobs:   protoJobs,
		More:          more,
		NextPageToken: next,
	}
	return res, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (s *Service) StorageJobsSummary(ctx context.Context, req *userPb.StorageJobsSummaryRequest) (*userPb.StorageJobsSummaryResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
	}

	queuedJobs, _, _, err := i.ListStorageJobs(api.ListStorageJobsConfig{Select: api.Queued, CidFilter: c})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing queued jobs: %v", err)
	}
	executingJobs, _, _, err := i.ListStorageJobs(api.ListStorageJobsConfig{Select: api.Executing, CidFilter: c})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing executing jobs: %v", err)
	}
	finalJobs, _, _, err := i.ListStorageJobs(api.ListStorageJobsConfig{Select: api.Final, CidFilter: c})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing final jobs: %v", err)
	}

	var queuedJobIDs []string
	for _, job := range queuedJobs {
		queuedJobIDs = append(queuedJobIDs, job.ID.String())
	}
	var executingJobIDs []string
	for _, job := range executingJobs {
		executingJobIDs = append(executingJobIDs, job.ID.String())
	}
	var finalJobIDs []string
	for _, job := range finalJobs {
		finalJobIDs = append(finalJobIDs, job.ID.String())
	}

	return &userPb.StorageJobsSummaryResponse{
		QueuedStorageJobs:    queuedJobIDs,
		ExecutingStorageJobs: executingJobIDs,
		FinalStorageJobs:     finalJobIDs,
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
		rpcJob, err := su.ToRPCJob(job)
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
