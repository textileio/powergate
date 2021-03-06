package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// StorageInfo provides access to Powergate storage indo APIs.
type StorageInfo struct {
	client adminPb.AdminServiceClient
}

// Get returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns an error with codes.NotFound.
func (s *StorageInfo) Get(ctx context.Context, userID, cid string) (*adminPb.StorageInfoResponse, error) {
	return s.client.StorageInfo(ctx, &adminPb.StorageInfoRequest{UserId: userID, Cid: cid})
}

// List returns a list of information about all stored cids, filtered by user ids and cids if provided.
func (s *StorageInfo) List(ctx context.Context, userIDs, cids []string) (*adminPb.ListStorageInfoResponse, error) {
	return s.client.ListStorageInfo(ctx, &adminPb.ListStorageInfoRequest{UserIds: userIDs, Cids: cids})
}
