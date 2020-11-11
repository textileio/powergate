package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/api/gen/powergate/admin/v1"
)

// StorageJobs provides access to Powergate jobs admin APIs.
type StorageJobs struct {
	client adminPb.AdminServiceClient
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
