package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	client proto.PowergateAdminServiceClient
}

// CreateInstance creates a new FFS instance, returning the instance ID and auth token.
func (a *Admin) CreateInstance(ctx context.Context) (*proto.CreateInstanceResponse, error) {
	return a.client.CreateInstance(ctx, &proto.CreateInstanceRequest{})
}

// ListInstances returns a list of existing API instances.
func (a *Admin) ListInstances(ctx context.Context) (*proto.ListInstancesResponse, error) {
	return a.client.ListInstances(ctx, &proto.ListInstancesRequest{})
}

// GetQueuedStorageJobs returns a list of queued storage jobs.
func (a *Admin) GetQueuedStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.GetQueuedStorageJobsResponse, error) {
	req := &proto.GetQueuedStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.GetQueuedStorageJobs(ctx, req)
}

// GetExecutingStorageJobs returns a list of executing storage jobs.
func (a *Admin) GetExecutingStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.GetExecutingStorageJobsResponse, error) {
	req := &proto.GetExecutingStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.GetExecutingStorageJobs(ctx, req)
}

// GetLatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *Admin) GetLatestFinalStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.GetLatestFinalStorageJobsResponse, error) {
	req := &proto.GetLatestFinalStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.GetLatestFinalStorageJobs(ctx, req)
}

// GetLatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *Admin) GetLatestSuccessfulStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.GetLatestSuccessfulStorageJobsResponse, error) {
	req := &proto.GetLatestSuccessfulStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.GetLatestSuccessfulStorageJobs(ctx, req)
}
