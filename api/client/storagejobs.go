package client

import (
	"context"
	"io"

	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageJobs provides access to Powergate jobs APIs.
type StorageJobs struct {
	client userPb.UserServiceClient
}

// ListSelect specifies which StorageJobs to list.
type ListSelect int32

const (
	// All lists all StorageJobs and is the default.
	All ListSelect = iota
	// Queued lists queued StorageJobs.
	Queued
	// Executing lists executing StorageJobs.
	Executing
	// Final lists final StorageJobs.
	Final
)

// ListConfig controls the behavior for listing StorageJobs.
type ListConfig struct {
	// CidFilter filters StorageJobs list to the specified cid. Defaults to no filter.
	CidFilter string
	// Limit limits the number of StorageJobs returned. Defaults to no limit.
	Limit uint64
	// Ascending returns the StorageJobs ascending by time. Defaults to false, descending.
	Ascending bool
	// Select specifies to return StorageJobs in the specified state.
	Select ListSelect
	// NextPageToken sets the slug from which to start building the next page of results.
	NextPageToken string
}

type summaryConfig struct {
	cid string
}

// SummaryOption configures a storageJobsConfig.
type SummaryOption = func(*summaryConfig)

// WithCid filters the results to the specified data cid.
func WithCid(cid string) SummaryOption {
	return func(conf *summaryConfig) {
		conf.cid = cid
	}
}

// WatchStorageJobsEvent represents an event for Watching a job.
type WatchStorageJobsEvent struct {
	Res *userPb.WatchStorageJobsResponse
	Err error
}

// Get returns the current state of the specified job.
func (j *StorageJobs) Get(ctx context.Context, jobID string) (*userPb.StorageJobResponse, error) {
	return j.client.StorageJob(ctx, &userPb.StorageJobRequest{JobId: jobID})
}

// StorageConfig returns the StorageConfig associated with the specified job.
func (j *StorageJobs) StorageConfig(ctx context.Context, jobID string) (*userPb.StorageConfigForJobResponse, error) {
	return j.client.StorageConfigForJob(ctx, &userPb.StorageConfigForJobRequest{JobId: jobID})
}

// List lists StorageJobs according to the provided ListConfig.
func (j *StorageJobs) List(ctx context.Context, config ListConfig) (*userPb.ListStorageJobsResponse, error) {
	sel := userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_UNSPECIFIED
	switch config.Select {
	case All:
		sel = userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_ALL
	case Queued:
		sel = userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_QUEUED
	case Executing:
		sel = userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_EXECUTING
	case Final:
		sel = userPb.StorageJobsSelector_STORAGE_JOBS_SELECTOR_FINAL
	}
	req := &userPb.ListStorageJobsRequest{
		CidFilter:     config.CidFilter,
		Limit:         config.Limit,
		Ascending:     config.Ascending,
		NextPageToken: config.NextPageToken,
		Selector:      sel,
	}
	return j.client.ListStorageJobs(ctx, req)
}

// Summary returns a summary of storage jobs.
func (j *StorageJobs) Summary(ctx context.Context, opts ...SummaryOption) (*userPb.StorageJobsSummaryResponse, error) {
	conf := &summaryConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &userPb.StorageJobsSummaryRequest{
		Cid: conf.cid,
	}
	return j.client.StorageJobsSummary(ctx, req)
}

// Watch pushes JobEvents to the provided channel. The provided channel will be owned
// by the client after the call, so it shouldn't be closed by the client. To stop receiving
// events, the provided ctx should be canceled. If an error occurs, it will be returned
// in the Err field of JobEvent and the channel will be closed.
func (j *StorageJobs) Watch(ctx context.Context, ch chan<- WatchStorageJobsEvent, jobIDs ...string) error {
	stream, err := j.client.WatchStorageJobs(ctx, &userPb.WatchStorageJobsRequest{JobIds: jobIDs})
	if err != nil {
		return err
	}
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- WatchStorageJobsEvent{Err: err}
				close(ch)
				break
			}
			ch <- WatchStorageJobsEvent{Res: res}
		}
	}()
	return nil
}

// Cancel signals that the executing Job with JobID jid should be
// canceled.
func (j *StorageJobs) Cancel(ctx context.Context, jobID string) (*userPb.CancelStorageJobResponse, error) {
	return j.client.CancelStorageJob(ctx, &userPb.CancelStorageJobRequest{JobId: jobID})
}
