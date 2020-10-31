package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// StorageJobs provides access to Powergate jobs admin APIs.
type StorageJobs struct {
	client proto.PowergateAdminServiceClient
}

type storageJobsConfig struct {
	ProfileID string
	Cids      []string
}

// StorageJobsOption configures a storageJobsConfig.
type StorageJobsOption = func(*storageJobsConfig)

// WithProfileID filters the results to the specified profile.
func WithProfileID(profileID string) StorageJobsOption {
	return func(conf *storageJobsConfig) {
		conf.ProfileID = profileID
	}
}

// WithCids filters the results to the specified data cids.
func WithCids(cids ...string) StorageJobsOption {
	return func(conf *storageJobsConfig) {
		conf.Cids = cids
	}
}

// Queued returns a list of queued storage jobs.
func (j *StorageJobs) Queued(ctx context.Context, opts ...StorageJobsOption) (*proto.QueuedStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &proto.QueuedStorageJobsRequest{
		ProfileId: conf.ProfileID,
		Cids:      conf.Cids,
	}
	return j.client.QueuedStorageJobs(ctx, req)
}

// Executing returns a list of executing storage jobs.
func (j *StorageJobs) Executing(ctx context.Context, opts ...StorageJobsOption) (*proto.ExecutingStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &proto.ExecutingStorageJobsRequest{
		ProfileId: conf.ProfileID,
		Cids:      conf.Cids,
	}
	return j.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinal returns a list of latest final storage jobs.
func (j *StorageJobs) LatestFinal(ctx context.Context, opts ...StorageJobsOption) (*proto.LatestFinalStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &proto.LatestFinalStorageJobsRequest{
		ProfileId: conf.ProfileID,
		Cids:      conf.Cids,
	}
	return j.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessful returns a list of latest successful storage jobs.
func (j *StorageJobs) LatestSuccessful(ctx context.Context, opts ...StorageJobsOption) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &proto.LatestSuccessfulStorageJobsRequest{
		ProfileId: conf.ProfileID,
		Cids:      conf.Cids,
	}
	return j.client.LatestSuccessfulStorageJobs(ctx, req)
}

// Summary returns a summary of storage jobs.
func (j *StorageJobs) Summary(ctx context.Context, opts ...StorageJobsOption) (*proto.StorageJobsSummaryResponse, error) {
	conf := &storageJobsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	req := &proto.StorageJobsSummaryRequest{
		ProfileId: conf.ProfileID,
		Cids:      conf.Cids,
	}
	return j.client.StorageJobsSummary(ctx, req)
}
