package cmd

import (
	"context"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCmd.AddCommand(ffsIDCmd)
}

var ffsIDCmd = &cobra.Command{
	Use:   "id",
	Short: "Returns the FFS instance id",
	Long:  `Returns the FFS instance id`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		s := spin.New("%s Getting FFS instance id...")
		s.Start()
		id, err := fcClient.FFS.ID(mustAuthCtx(ctx))
		s.Stop()
		checkErr(err)
		Message("FFS instance id: %s", id.String())
	},
}
