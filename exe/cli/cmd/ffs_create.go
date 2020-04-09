package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCmd.AddCommand(ffsCreateCmd)
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
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Creating ffs instance...")
		s.Start()
		id, token, err := fcClient.Ffs.Create(ctx)
		s.Stop()
		checkErr(err)
		Message("Instance created with id %s and token %s", id, token)

	},
}
