package cmd

import (
	"context"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCloseCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsCloseCmd)
}

var ffsCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "Close the FFS instance",
	Long:  `Close the FFS instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		s := spin.New("%s Closing FFS instance...")
		s.Start()
		err := fcClient.FFS.Close(authCtx(ctx))
		s.Stop()
		checkErr(err)
		Message("FFS instance closed")
	},
}
