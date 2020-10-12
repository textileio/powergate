package cmd

import (
	"context"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ffsCreateInstanceTimeout = time.Second * 30
)

func init() {
	ffsCmd.AddCommand(ffsCreateCmd)
	ffsCreateCmd.Flags().StringP("token", "t", "", "FFS admin token (if applies)")
}

var ffsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create ffs instance",
	Long:  `Create ffs instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), ffsCreateInstanceTimeout)
		defer cancel()

		s := spin.New("%s Creating ffs instance...")
		s.Start()
		res, err := fcClient.Admin.CreateInstance(ctx)
		s.Stop()
		checkErr(err)
		Message("Instance created with id %s and token %s", res.Id, res.Token)
	},
}
