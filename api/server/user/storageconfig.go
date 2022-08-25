package user

import (
	"context"

	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/api"
	"github.com/textileio/powergate/v2/util"
)

// DefaultStorageConfig calls ffs.DefaultStorageConfig.
func (s *Service) DefaultStorageConfig(ctx context.Context, req *userPb.DefaultStorageConfigRequest) (*userPb.DefaultStorageConfigResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	conf := i.DefaultStorageConfig()
	return &userPb.DefaultStorageConfigResponse{
		DefaultStorageConfig: toRPCStorageConfig(conf),
	}, nil
}

// SetDefaultStorageConfig sets a new config to be used by default.
func (s *Service) SetDefaultStorageConfig(ctx context.Context, req *userPb.SetDefaultStorageConfigRequest) (*userPb.SetDefaultStorageConfigResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	defaultConfig := ffs.StorageConfig{
		Repairable:    req.Config.Repairable,
		Hot:           fromRPCHotConfig(req.Config.Hot),
		Cold:          fromRPCColdConfig(req.Config.Cold),
		Notifications: fromRPCNotificationConfigs(req.Config.Notifications),
	}
	if err := i.SetDefaultStorageConfig(defaultConfig); err != nil {
		return nil, err
	}
	return &userPb.SetDefaultStorageConfigResponse{}, nil
}

// ApplyStorageConfig applies the provided cid storage config.
func (s *Service) ApplyStorageConfig(ctx context.Context, req *userPb.ApplyStorageConfigRequest) (*userPb.ApplyStorageConfigResponse, error) {
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
			Repairable:    req.Config.Repairable,
			Hot:           fromRPCHotConfig(req.Config.Hot),
			Cold:          fromRPCColdConfig(req.Config.Cold),
			Notifications: fromRPCNotificationConfigs(req.Config.Notifications),
		}
		options = append(options, api.WithStorageConfig(config))
	}

	if req.HasOverrideConfig {
		options = append(options, api.WithOverride(req.OverrideConfig))
	}

	if len(req.ImportDealIds) > 0 {
		options = append(options, api.WithDealImport(req.ImportDealIds))
	}
	if req.NoExec {
		options = append(options, api.WithNoExec(true))
	}

	jid, err := i.PushStorageConfig(c, options...)
	if err != nil {
		return nil, err
	}

	return &userPb.ApplyStorageConfigResponse{
		JobId: jid.String(),
	}, nil
}

// Remove calls ffs.Remove.
func (s *Service) Remove(ctx context.Context, req *userPb.RemoveRequest) (*userPb.RemoveResponse, error) {
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

	return &userPb.RemoveResponse{}, nil
}
