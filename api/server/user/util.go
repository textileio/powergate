package user

import (
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
)

func toRPCStorageConfig(config ffs.StorageConfig) *userPb.StorageConfig {
	return &userPb.StorageConfig{
		Repairable: config.Repairable,
		Hot:        toRPCHotConfig(config.Hot),
		Cold:       toRPCColdConfig(config.Cold),
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
		},
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
