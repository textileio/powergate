package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// StorageJobs provides access to Powergate jobs admin APIs.
type StorageJobs struct {
	client adminPb.AdminServiceClient
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
	// UserIDFilter filters StorageJobs list to the specified user ID. Defaults to no filter.
	UserIDFilter string
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

type storageJobsConfig struct {
	UserID string
	Cids   []string
}

// StorageJobsOption configures a storageJobsConfig.
type StorageJobsOption = func(*storageJobsConfig)

// WithUserID filters the results to the specified user.
func WithUserID(userID string) StorageJobsOption {
	return func(conf *storageJobsConfig) {
		conf.UserID = userID
	}
}

// WithCids filters the results to the specified data cids.
func WithCids(cids ...string) StorageJobsOption {
	return func(conf *storageJobsConfig) {
		conf.Cids = cids
	}
}

// List lists StorageJobs according to the provided ListConfig.
func (j *StorageJobs) List(ctx context.Context, config ListConfig) (*adminPb.ListStorageJobsResponse, error) {
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
	req := &adminPb.ListStorageJobsRequest{
		CidFilter:     config.CidFilter,
		Limit:         config.Limit,
		Ascending:     config.Ascending,
		NextPageToken: config.NextPageToken,
		Selector:      sel,
	}
	return j.client.ListStorageJobs(ctx, req)
}

// Queued returns a list of queued storage jobs.
func (j *StorageJobs) Queued(ctx context.Context, opts ...StorageJobsOption) (*adminPb.QueuedStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &adminPb.QueuedStorageJobsRequest{
		UserId: conf.UserID,
		Cids:   conf.Cids,
	}
	return j.client.QueuedStorageJobs(ctx, req)
}

// Executing returns a list of executing storage jobs.
func (j *StorageJobs) Executing(ctx context.Context, opts ...StorageJobsOption) (*adminPb.ExecutingStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &adminPb.ExecutingStorageJobsRequest{
		UserId: conf.UserID,
		Cids:   conf.Cids,
	}
	return j.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinal returns a list of latest final storage jobs.
func (j *StorageJobs) LatestFinal(ctx context.Context, opts ...StorageJobsOption) (*adminPb.LatestFinalStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &adminPb.LatestFinalStorageJobsRequest{
		UserId: conf.UserID,
		Cids:   conf.Cids,
	}
	return j.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessful returns a list of latest successful storage jobs.
func (j *StorageJobs) LatestSuccessful(ctx context.Context, opts ...StorageJobsOption) (*adminPb.LatestSuccessfulStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &adminPb.LatestSuccessfulStorageJobsRequest{
		UserId: conf.UserID,
		Cids:   conf.Cids,
	}
	return j.client.LatestSuccessfulStorageJobs(ctx, req)
}

// Summary returns a summary of storage jobs.
func (j *StorageJobs) Summary(ctx context.Context, opts ...StorageJobsOption) (*adminPb.StorageJobsSummaryResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &adminPb.StorageJobsSummaryRequest{
		UserId: conf.UserID,
		Cids:   conf.Cids,
	}
	return j.client.StorageJobsSummary(ctx, req)
}
