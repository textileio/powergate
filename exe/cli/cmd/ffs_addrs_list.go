package cmd

import (
	"context"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsAddrsListCmd.Flags().StringP("token", "t", "", "token of the request")

	ffsAddrsCmd.AddCommand(ffsAddrsListCmd)
}

var ffsAddrsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the wallet adresses for the ffs instance",
	Long:  `List the wallet adresses for the ffs instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Retrieving addresses...")
		s.Start()
		addrs, err := fcClient.Ffs.Addrs(authCtx(ctx))
		checkErr(err)
		defaultConfig, err := fcClient.Ffs.DefaultConfig(authCtx(ctx))
		checkErr(err)
		s.Stop()

		data := make([][]string, len(addrs))
		for i, addr := range addrs {
			isDefault := ""
			if addr.Addr == defaultConfig.Cold.Filecoin.Addr {
				isDefault = "yes"
			}
			data[i] = []string{addr.Name, addr.Addr, isDefault}
		}
		Message("Wallet addresses:")
		RenderTable(os.Stdout, []string{"name", "address", "default"}, data)
	},
}
