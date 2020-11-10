package admin

import (
	"context"

	proto "github.com/textileio/powergate/proto/admin/v1"
)

// Profiles provides access to Powergate admin storage profile APIs.
type Profiles struct {
	client proto.PowergateAdminServiceClient
}

// Create creates a new Powergate storage profile, returning the instance ID and auth token.
func (p *Profiles) Create(ctx context.Context) (*proto.CreateStorageProfileResponse, error) {
	return p.client.CreateStorageProfile(ctx, &proto.CreateStorageProfileRequest{})
}

// List returns a list of existing API instances.
func (p *Profiles) List(ctx context.Context) (*proto.StorageProfilesResponse, error) {
	return p.client.StorageProfiles(ctx, &proto.StorageProfilesRequest{})
}
