package client

import (
	"context"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// StorageConfig provides access to Powergate storage config APIs.
type StorageConfig struct {
	client userPb.UserServiceClient
}

// ApplyOption mutates a push request.
type ApplyOption func(r *userPb.ApplyStorageConfigRequest)

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c *userPb.StorageConfig) ApplyOption {
	return func(r *userPb.ApplyStorageConfigRequest) {
		r.HasConfig = true
		r.Config = c
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) ApplyOption {
	return func(r *userPb.ApplyStorageConfigRequest) {
		r.HasOverrideConfig = true
		r.OverrideConfig = override
	}
}

// Default returns the default storage config.
func (s *StorageConfig) Default(ctx context.Context) (*userPb.DefaultStorageConfigResponse, error) {
	return s.client.DefaultStorageConfig(ctx, &userPb.DefaultStorageConfigRequest{})
}

// SetDefault sets the default storage config.
func (s *StorageConfig) SetDefault(ctx context.Context, config *userPb.StorageConfig) (*userPb.SetDefaultStorageConfigResponse, error) {
	req := &userPb.SetDefaultStorageConfigRequest{
		Config: config,
	}
	return s.client.SetDefaultStorageConfig(ctx, req)
}

// Apply push a new configuration for the Cid in the Hot and Cold layers.
func (s *StorageConfig) Apply(ctx context.Context, cid string, opts ...ApplyOption) (*userPb.ApplyStorageConfigResponse, error) {
	req := &userPb.ApplyStorageConfigRequest{Cid: cid}
	for _, opt := range opts {
		opt(req)
	}
	return s.client.ApplyStorageConfig(ctx, req)
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (s *StorageConfig) Remove(ctx context.Context, cid string) (*userPb.RemoveResponse, error) {
	return s.client.Remove(ctx, &userPb.RemoveRequest{Cid: cid})
}
