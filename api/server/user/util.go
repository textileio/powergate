package user

import (
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
)

func toRPCStorageConfig(config ffs.StorageConfig) *userPb.StorageConfig {
	return &userPb.StorageConfig{
		Repairable:    config.Repairable,
		Hot:           toRPCHotConfig(config.Hot),
		Cold:          toRPCColdConfig(config.Cold),
		Notifications: toRPCNotificationConfigs(config.Notifications),
	}
}

func toRPCHotConfig(config ffs.HotConfig) *userPb.HotConfig {
	return &userPb.HotConfig{
		Enabled:          config.Enabled,
		AllowUnfreeze:    config.AllowUnfreeze,
		UnfreezeMaxPrice: config.UnfreezeMaxPrice,
		Ipfs: &userPb.IpfsConfig{
			AddTimeout: int64(config.Ipfs.AddTimeout),
		},
	}
}

func toRPCColdConfig(config ffs.ColdConfig) *userPb.ColdConfig {
	return &userPb.ColdConfig{
		Enabled: config.Enabled,
		Filecoin: &userPb.FilConfig{
			ReplicationFactor: int64(config.Filecoin.RepFactor),
			DealMinDuration:   config.Filecoin.DealMinDuration,
			ExcludedMiners:    config.Filecoin.ExcludedMiners,
			TrustedMiners:     config.Filecoin.TrustedMiners,
			CountryCodes:      config.Filecoin.CountryCodes,
			Renew: &userPb.FilRenew{
				Enabled:   config.Filecoin.Renew.Enabled,
				Threshold: int64(config.Filecoin.Renew.Threshold),
			},
			Address:         config.Filecoin.Addr,
			MaxPrice:        config.Filecoin.MaxPrice,
			FastRetrieval:   config.Filecoin.FastRetrieval,
			DealStartOffset: config.Filecoin.DealStartOffset,
			VerifiedDeal:    config.Filecoin.VerifiedDeal,
		},
	}
}

func toRPCNotificationConfigs(notifications []*ffs.NotificationConfig) []*userPb.NotificationConfig {
	if notifications == nil {
		return nil
	}

	var out []*userPb.NotificationConfig

	for _, cfg := range notifications {
		if res := toRPCNotificationConfig(cfg); res != nil {
			out = append(out, res)
		}
	}

	return out
}

func toRPCNotificationConfig(cfg *ffs.NotificationConfig) *userPb.NotificationConfig {
	if cfg == nil {
		return nil
	}

	return &userPb.NotificationConfig{
		Webhook:       toRPCWebhook(cfg.Webhook),
		Configuration: toRPCWebhookConfiguration(cfg.Configuration),
	}
}

func toRPCWebhook(webhook *ffs.Webhook) *userPb.Webhook {
	if webhook == nil {
		return nil
	}

	return &userPb.Webhook{
		Endpoint:       webhook.Endpoint,
		Authentication: toRPCWebhookAuth(webhook.Authentication),
	}
}

func toRPCWebhookAuth(authentication *ffs.WebhookAuthentication) *userPb.WebhookAuthentication {
	if authentication == nil {
		return nil
	}

	return &userPb.WebhookAuthentication{
		Type: authentication.Type,
		Data: toRPCWebhookAuthData(authentication.Data),
	}
}

func toRPCWebhookAuthData(data *ffs.WebhookAuthData) *userPb.WebhookAuthData {
	if data == nil {
		return nil
	}

	return &userPb.WebhookAuthData{
		Username: data.Username,
		Password: data.Password,
	}
}

func toRPCWebhookConfiguration(configuration *ffs.WebhookConfiguration) *userPb.WebhookConfiguration {
	if configuration == nil {
		return nil
	}

	return &userPb.WebhookConfiguration{
		Events: configuration.Events,
		Alerts: toRPCWebhookAlerts(configuration.Alerts),
	}
}

func toRPCWebhookAlerts(alerts []*ffs.WebhookAlert) []*userPb.WebhookAlert {
	if alerts == nil {
		return nil
	}

	var out []*userPb.WebhookAlert
	for _, alert := range alerts {
		if res := toRPCWebhookAlert(alert); res != nil {
			out = append(out, res)
		}
	}

	return out
}

func toRPCWebhookAlert(alert *ffs.WebhookAlert) *userPb.WebhookAlert {
	if alert == nil {
		return nil
	}

	return &userPb.WebhookAlert{
		Type:      alert.Type,
		Threshold: alert.Threshold,
	}
}

func fromRPCHotConfig(config *userPb.HotConfig) ffs.HotConfig {
	res := ffs.HotConfig{}
	if config != nil {
		res.Enabled = config.Enabled
		res.AllowUnfreeze = config.AllowUnfreeze
		res.UnfreezeMaxPrice = config.UnfreezeMaxPrice
		if config.Ipfs != nil {
			ipfs := ffs.IpfsConfig{
				AddTimeout: int(config.Ipfs.AddTimeout),
			}
			res.Ipfs = ipfs
		}
	}
	return res
}

func fromRPCColdConfig(config *userPb.ColdConfig) ffs.ColdConfig {
	res := ffs.ColdConfig{}
	if config != nil {
		res.Enabled = config.Enabled
		if config.Filecoin != nil {
			filecoin := ffs.FilConfig{
				RepFactor:       int(config.Filecoin.ReplicationFactor),
				DealMinDuration: config.Filecoin.DealMinDuration,
				ExcludedMiners:  config.Filecoin.ExcludedMiners,
				CountryCodes:    config.Filecoin.CountryCodes,
				TrustedMiners:   config.Filecoin.TrustedMiners,
				Addr:            config.Filecoin.Address,
				MaxPrice:        config.Filecoin.MaxPrice,
				FastRetrieval:   config.Filecoin.FastRetrieval,
				DealStartOffset: config.Filecoin.DealStartOffset,
				VerifiedDeal:    config.Filecoin.VerifiedDeal,
			}
			if config.Filecoin.Renew != nil {
				renew := ffs.FilRenew{
					Enabled:   config.Filecoin.Renew.Enabled,
					Threshold: int(config.Filecoin.Renew.Threshold),
				}
				filecoin.Renew = renew
			}
			res.Filecoin = filecoin
		}
	}
	return res
}

func fromRPCNotificationConfigs(configs []*userPb.NotificationConfig) []*ffs.NotificationConfig {
	if configs == nil {
		return nil
	}

	var out []*ffs.NotificationConfig
	for _, cfg := range configs {
		res := fromRPCNotificationConfig(cfg)

		if res != nil {
			out = append(out, res)
		}
	}

	return out
}

func fromRPCNotificationConfig(config *userPb.NotificationConfig) *ffs.NotificationConfig {
	if config == nil {
		return nil
	}

	return &ffs.NotificationConfig{
		Webhook:       fromRPCWebhook(config.Webhook),
		Configuration: fromRPCWebhookConfiguration(config.Configuration),
	}
}

func fromRPCWebhookConfiguration(configuration *userPb.WebhookConfiguration) *ffs.WebhookConfiguration {
	if configuration == nil {
		return nil
	}

	return &ffs.WebhookConfiguration{
		Events: configuration.Events,
		Alerts: fromRPCWebhookAlerts(configuration.Alerts),
	}
}

func fromRPCWebhookAlerts(alerts []*userPb.WebhookAlert) []*ffs.WebhookAlert {
	if alerts == nil {
		return nil
	}

	var out []*ffs.WebhookAlert
	for _, alert := range alerts {
		res := fromRPCWebhookAlert(alert)
		if res != nil {
			out = append(out, res)
		}
	}

	return out
}

func fromRPCWebhookAlert(alert *userPb.WebhookAlert) *ffs.WebhookAlert {
	if alert == nil {
		return nil
	}

	return &ffs.WebhookAlert{
		Type:      alert.Type,
		Threshold: alert.Threshold,
	}
}

func fromRPCWebhook(webhook *userPb.Webhook) *ffs.Webhook {
	if webhook == nil {
		return nil
	}

	return &ffs.Webhook{
		Endpoint:       webhook.Endpoint,
		Authentication: fromRPCWebhookAuthentication(webhook.Authentication),
	}
}

func fromRPCWebhookAuthentication(authentication *userPb.WebhookAuthentication) *ffs.WebhookAuthentication {
	if authentication == nil {
		return nil
	}

	return &ffs.WebhookAuthentication{
		Type: authentication.Type,
		Data: fromRPCWebhookAuthenticationData(authentication.Data),
	}
}

func fromRPCWebhookAuthenticationData(data *userPb.WebhookAuthData) *ffs.WebhookAuthData {
	if data == nil {
		return nil
	}

	return &ffs.WebhookAuthData{
		Username: data.Username,
		Password: data.Password,
	}
}

func buildListDealRecordsOptions(conf *userPb.DealRecordsConfig) []deals.DealRecordsOption {
	var opts []deals.DealRecordsOption
	if conf != nil {
		opts = []deals.DealRecordsOption{
			deals.WithAscending(conf.Ascending),
			deals.WithDataCids(conf.DataCids...),
			deals.WithFromAddrs(conf.FromAddrs...),
			deals.WithIncludePending(conf.IncludePending),
			deals.WithIncludeFinal(conf.IncludeFinal),
			deals.WithIncludeFailed(conf.IncludeFailed),
		}
	}
	return opts
}
