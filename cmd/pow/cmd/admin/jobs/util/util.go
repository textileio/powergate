package util

import (
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client/admin"
)

// StorageJobsOpts returns storage jobs query options from viper config.
func StorageJobsOpts() []admin.StorageJobsOption {
	var opts []admin.StorageJobsOption
	if viper.IsSet("user-id") {
		opts = append(opts, admin.WithUserID(viper.GetString("user-id")))
	}
	if viper.IsSet("cids") {
		opts = append(opts, admin.WithCids(viper.GetStringSlice("cids")...))
	}
	return opts
}
