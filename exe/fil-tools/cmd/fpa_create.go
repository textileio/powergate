package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpaCmd.AddCommand(fpaCreateCmd)
}

var fpaCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create fpa instance",
	Long:  `Create fpa instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Creating fpa instance...")
		s.Start()
		id, token, err := fcClient.Fpa.Create(ctx)
		checkErr(err)
		s.Stop()
		Message("Instance created with id %s and token %s.", id, token)

	},
}
