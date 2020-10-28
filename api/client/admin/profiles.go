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

// ListStorageProfiles returns a list of existing API instances.
func (p *Profiles) ListStorageProfiles(ctx context.Context) (*proto.ListStorageProfilesResponse, error) {
	return p.client.ListStorageProfiles(ctx, &proto.ListStorageProfilesRequest{})
}
