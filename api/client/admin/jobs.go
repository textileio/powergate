package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Jobs provides access to Powergate jobs admin APIs.
type Jobs struct {
	client proto.PowergateAdminServiceClient
}

// QueuedStorageJobs returns a list of queued storage jobs.
func (j *Jobs) QueuedStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.QueuedStorageJobsResponse, error) {
	req := &proto.QueuedStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return j.client.QueuedStorageJobs(ctx, req)
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (j *Jobs) ExecutingStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.ExecutingStorageJobsResponse, error) {
	req := &proto.ExecutingStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return j.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (j *Jobs) LatestFinalStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.LatestFinalStorageJobsResponse, error) {
	req := &proto.LatestFinalStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return j.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (j *Jobs) LatestSuccessfulStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	req := &proto.LatestSuccessfulStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return j.client.LatestSuccessfulStorageJobs(ctx, req)
}

// StorageJobsSummary returns a summary of storage jobs.
func (j *Jobs) StorageJobsSummary(ctx context.Context, instanceID string, cids ...string) (*proto.StorageJobsSummaryResponse, error) {
	req := &proto.StorageJobsSummaryRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return j.client.StorageJobsSummary(ctx, req)
}
