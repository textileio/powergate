package admin

import (
	"context"

	"github.com/ipfs/go-cid"
	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/api/server/util"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/scheduler"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListStorageJobs lists StorageJobs according to the provided request parameters.
func (a *Service) ListStorageJobs(ctx context.Context, req *adminPb.ListStorageJobsRequest) (*adminPb.ListStorageJobsResponse, error) {
	var selector scheduler.Select
	switch req.Selector {
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_ALL:
		selector = scheduler.All
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_EXECUTING:
		selector = scheduler.Executing
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_FINAL:
		selector = scheduler.Final
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_QUEUED:
		selector = scheduler.Queued
	case userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_UNSPECIFIED:
		selector = scheduler.All
	}
	conf := scheduler.ListStorageJobsConfig{
		APIIDFilter:   ffs.APIID(req.UserIdFilter),
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
	jobs, more, next, err := a.s.ListStorageJobs(conf)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing storage jobs: %v", err)
	}
	protoJobs, err := su.ToProtoStorageJobs(jobs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "converting jobs to protos: %v", err)
	}
	res := &adminPb.ListStorageJobsResponse{
		StorageJobs:   protoJobs,
		More:          more,
		NextPageToken: next,
	}
	return res, nil
}

// StorageJobsSummary returns a summary of all storage jobs.
func (a *Service) StorageJobsSummary(ctx context.Context, req *adminPb.StorageJobsSummaryRequest) (*adminPb.StorageJobsSummaryResponse, error) {
	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
	}

	queuedJobs, _, _, err := a.s.ListStorageJobs(scheduler.ListStorageJobsConfig{Select: scheduler.Queued, APIIDFilter: ffs.APIID(req.UserId), CidFilter: c})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing queued jobs: %v", err)
	}
	executingJobs, _, _, err := a.s.ListStorageJobs(scheduler.ListStorageJobsConfig{Select: scheduler.Executing, APIIDFilter: ffs.APIID(req.UserId), CidFilter: c})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing executing jobs: %v", err)
	}
	finalJobs, _, _, err := a.s.ListStorageJobs(scheduler.ListStorageJobsConfig{Select: scheduler.Final, APIIDFilter: ffs.APIID(req.UserId), CidFilter: c})
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

	return &adminPb.StorageJobsSummaryResponse{
		QueuedStorageJobs:    queuedJobIDs,
		ExecutingStorageJobs: executingJobIDs,
		FinalStorageJobs:     finalJobIDs,
	}, nil
}
