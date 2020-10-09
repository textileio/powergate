package cmd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsConfigCmd.AddCommand(ffsConfigDefaultCmd)
}

var ffsConfigDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Returns the default storage config",
	Long:  `Returns the default storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		s := spin.New("%s Getting default storage config...")
		s.Start()
		config, err := fcClient.FFS.DefaultStorageConfig(authCtx(ctx))
		s.Stop()
		checkErr(err)

		json, err := json.MarshalIndent(config, "", "  ")
		checkErr(err)

		Message("Default storage config:\n%s", string(json))
	},
}
