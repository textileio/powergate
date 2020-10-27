package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// StorageConfig provides access to Powergate storage config APIs.
type StorageConfig struct {
	client proto.PowergateServiceClient
}

// ApplyOption mutates a push request.
type ApplyOption func(r *proto.ApplyStorageConfigRequest)

// WithStorageConfig overrides the Api default Cid configuration.
func WithStorageConfig(c *proto.StorageConfig) ApplyOption {
	return func(r *proto.ApplyStorageConfigRequest) {
		r.HasConfig = true
		r.Config = c
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) ApplyOption {
	return func(r *proto.ApplyStorageConfigRequest) {
		r.HasOverrideConfig = true
		r.OverrideConfig = override
	}
}

// Default returns the default storage config.
func (s *StorageConfig) Default(ctx context.Context) (*proto.DefaultStorageConfigResponse, error) {
	return s.client.DefaultStorageConfig(ctx, &proto.DefaultStorageConfigRequest{})
}

// SetDefault sets the default storage config.
func (s *StorageConfig) SetDefault(ctx context.Context, config *proto.StorageConfig) (*proto.SetDefaultStorageConfigResponse, error) {
	req := &proto.SetDefaultStorageConfigRequest{
		Config: config,
	}
	return s.client.SetDefaultStorageConfig(ctx, req)
}

// Apply push a new configuration for the Cid in the Hot and Cold layers.
func (s *StorageConfig) Apply(ctx context.Context, cid string, opts ...ApplyOption) (*proto.ApplyStorageConfigResponse, error) {
	req := &proto.ApplyStorageConfigRequest{Cid: cid}
	for _, opt := range opts {
		opt(req)
	}
	return s.client.ApplyStorageConfig(ctx, req)
}

// Remove removes a Cid from being tracked as an active storage. The Cid should have
// both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage.
func (s *StorageConfig) Remove(ctx context.Context, cid string) (*proto.RemoveResponse, error) {
	return s.client.Remove(ctx, &proto.RemoveRequest{Cid: cid})
}
