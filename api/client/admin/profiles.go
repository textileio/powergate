package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Profiles provides access to Powergate admin storage profile APIs.
type Profiles struct {
	client proto.PowergateAdminServiceClient
}

// CreateStorageProfile creates a new Powergate storage profile, returning the instance ID and auth token.
func (p *Profiles) CreateStorageProfile(ctx context.Context) (*proto.CreateStorageProfileResponse, error) {
	return p.client.CreateStorageProfile(ctx, &proto.CreateStorageProfileRequest{})
}

// StorageProfiles returns a list of existing API instances.
func (p *Profiles) StorageProfiles(ctx context.Context) (*proto.StorageProfilesResponse, error) {
	return p.client.StorageProfiles(ctx, &proto.StorageProfilesRequest{})
}
