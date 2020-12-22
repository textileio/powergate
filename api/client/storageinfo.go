package client

import (
	"context"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
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

// QueryStorageInfo returns a list of infomration about all stored cids, filtered by cids if provided.
func (s *StorageInfo) QueryStorageInfo(ctx context.Context, cids ...string) (*userPb.QueryStorageInfoResponse, error) {
	return s.client.QueryStorageInfo(ctx, &userPb.QueryStorageInfoRequest{Cids: cids})
}
