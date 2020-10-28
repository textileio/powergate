package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// StorageJobs provides access to Powergate jobs admin APIs.
type StorageJobs struct {
	client proto.PowergateAdminServiceClient
}

// Queued returns a list of queued storage jobs.
func (j *StorageJobs) Queued(ctx context.Context, profileID string, cids ...string) (*proto.QueuedStorageJobsResponse, error) {
	req := &proto.QueuedStorageJobsRequest{
		ProfileId: profileID,
		Cids:      cids,
	}
	return j.client.QueuedStorageJobs(ctx, req)
}

// Executing returns a list of executing storage jobs.
func (j *StorageJobs) Executing(ctx context.Context, profileID string, cids ...string) (*proto.ExecutingStorageJobsResponse, error) {
	req := &proto.ExecutingStorageJobsRequest{
		ProfileId: profileID,
		Cids:      cids,
	}
	return j.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinal returns a list of latest final storage jobs.
func (j *StorageJobs) LatestFinal(ctx context.Context, profileID string, cids ...string) (*proto.LatestFinalStorageJobsResponse, error) {
	req := &proto.LatestFinalStorageJobsRequest{
		ProfileId: profileID,
		Cids:      cids,
	}
	return j.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessful returns a list of latest successful storage jobs.
func (j *StorageJobs) LatestSuccessful(ctx context.Context, profileID string, cids ...string) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	req := &proto.LatestSuccessfulStorageJobsRequest{
		ProfileId: profileID,
		Cids:      cids,
	}
	return j.client.LatestSuccessfulStorageJobs(ctx, req)
}

// Summary returns a summary of storage jobs.
func (j *StorageJobs) Summary(ctx context.Context, profileID string, cids ...string) (*proto.StorageJobsSummaryResponse, error) {
	req := &proto.StorageJobsSummaryRequest{
		ProfileId: profileID,
		Cids:      cids,
	}
	return j.client.StorageJobsSummary(ctx, req)
}
