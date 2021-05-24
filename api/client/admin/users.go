package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Users provides access to Powergate admin users APIs.
type Users struct {
	client adminPb.AdminServiceClient
}

// Create creates a new Powergate user, returning the user ID and auth token.
func (p *Users) Create(ctx context.Context) (*adminPb.CreateUserResponse, error) {
	return p.client.CreateUser(ctx, &adminPb.CreateUserRequest{})
}

// RegenerateAuth invalidates an existing token replacing it with a new one.
func (p *Users) RegenerateAuth(ctx context.Context, token string) (*adminPb.RegenerateAuthResponse, error) {
	return p.client.RegenerateAuth(ctx, &adminPb.RegenerateAuthRequest{Token: token})
}

// List returns a list of existing users.
func (p *Users) List(ctx context.Context) (*adminPb.UsersResponse, error) {
	return p.client.Users(ctx, &adminPb.UsersRequest{})
}
