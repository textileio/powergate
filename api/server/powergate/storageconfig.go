package powergate

import (
	"context"

	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"github.com/textileio/powergate/util"
)

// DefaultStorageConfig calls ffs.DefaultStorageConfig.
func (s *Service) DefaultStorageConfig(ctx context.Context, req *proto.DefaultStorageConfigRequest) (*proto.DefaultStorageConfigResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	conf := i.DefaultStorageConfig()
	return &proto.DefaultStorageConfigResponse{
		DefaultStorageConfig: ToRPCStorageConfig(conf),
	}, nil
}

// SetDefaultStorageConfig sets a new config to be used by default.
func (s *Service) SetDefaultStorageConfig(ctx context.Context, req *proto.SetDefaultStorageConfigRequest) (*proto.SetDefaultStorageConfigResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	defaultConfig := ffs.StorageConfig{
		Repairable: req.Config.Repairable,
		Hot:        fromRPCHotConfig(req.Config.Hot),
		Cold:       fromRPCColdConfig(req.Config.Cold),
	}
	if err := i.SetDefaultStorageConfig(defaultConfig); err != nil {
		return nil, err
	}
	return &proto.SetDefaultStorageConfigResponse{}, nil
}

// ApplyStorageConfig applies the provided cid storage config.
func (s *Service) ApplyStorageConfig(ctx context.Context, req *proto.ApplyStorageConfigRequest) (*proto.ApplyStorageConfigResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, err
	}

	options := []api.PushStorageConfigOption{}

	if req.HasConfig {
		config := ffs.StorageConfig{
			Repairable: req.Config.Repairable,
			Hot:        fromRPCHotConfig(req.Config.Hot),
			Cold:       fromRPCColdConfig(req.Config.Cold),
		}
		options = append(options, api.WithStorageConfig(config))
	}

	if req.HasOverrideConfig {
		options = append(options, api.WithOverride(req.OverrideConfig))
	}

	jid, err := i.PushStorageConfig(c, options...)
	if err != nil {
		return nil, err
	}

	return &proto.ApplyStorageConfigResponse{
		JobId: jid.String(),
	}, nil
}

// Remove calls ffs.Remove.
func (s *Service) Remove(ctx context.Context, req *proto.RemoveRequest) (*proto.RemoveResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, err
	}

	if err := i.Remove(c); err != nil {
		return nil, err
	}

	return &proto.RemoveResponse{}, nil
}
