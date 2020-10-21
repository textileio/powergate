package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Admin provides access to Powergate admin APIs.
type Admin struct {
	client proto.PowergateAdminServiceClient
}

// CreateStorageProfile creates a new Powergate storage profile, returning the instance ID and auth token.
func (a *Admin) CreateStorageProfile(ctx context.Context) (*proto.CreateStorageProfileResponse, error) {
	return a.client.CreateStorageProfile(ctx, &proto.CreateStorageProfileRequest{})
}

// ListStorageProfiles returns a list of existing API instances.
func (a *Admin) ListStorageProfiles(ctx context.Context) (*proto.ListStorageProfilesResponse, error) {
	return a.client.ListStorageProfiles(ctx, &proto.ListStorageProfilesRequest{})
}

// QueuedStorageJobs returns a list of queued storage jobs.
func (a *Admin) QueuedStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.QueuedStorageJobsResponse, error) {
	req := &proto.QueuedStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.QueuedStorageJobs(ctx, req)
}

// ExecutingStorageJobs returns a list of executing storage jobs.
func (a *Admin) ExecutingStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.ExecutingStorageJobsResponse, error) {
	req := &proto.ExecutingStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinalStorageJobs returns a list of latest final storage jobs.
func (a *Admin) LatestFinalStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.LatestFinalStorageJobsResponse, error) {
	req := &proto.LatestFinalStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessfulStorageJobs returns a list of latest successful storage jobs.
func (a *Admin) LatestSuccessfulStorageJobs(ctx context.Context, instanceID string, cids ...string) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	req := &proto.LatestSuccessfulStorageJobsRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.LatestSuccessfulStorageJobs(ctx, req)
}

// StorageJobsSummary returns a summary of storage jobs.
func (a *Admin) StorageJobsSummary(ctx context.Context, instanceID string, cids ...string) (*proto.StorageJobsSummaryResponse, error) {
	req := &proto.StorageJobsSummaryRequest{
		InstanceId: instanceID,
		Cids:       cids,
	}
	return a.client.StorageJobsSummary(ctx, req)
}
