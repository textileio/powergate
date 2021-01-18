package client

import (
	"context"

	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
)

// StorageInfo provides access to Powergate storage indo APIs.
type StorageInfo struct {
	client userPb.UserServiceClient
}

// StorageInfo returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns an error with codes.NotFound.
func (s *StorageInfo) StorageInfo(ctx context.Context, cid string) (*userPb.StorageInfoResponse, error) {
	return s.client.StorageInfo(ctx, &userPb.StorageInfoRequest{Cid: cid})
}

// ListStorageInfo returns a list of information about all stored cids, filtered by cids if provided.
func (s *StorageInfo) ListStorageInfo(ctx context.Context, cids ...string) (*userPb.ListStorageInfoResponse, error) {
	return s.client.ListStorageInfo(ctx, &userPb.ListStorageInfoRequest{Cids: cids})
}
