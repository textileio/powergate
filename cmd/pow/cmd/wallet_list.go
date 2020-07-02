package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	walletCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all wallet addresses",
	Long:  `Print all wallet addresses`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New(fmt.Sprintf("%s Getting wallet addresses...", "%s"))
		s.Start()
		addrs, err := fcClient.Wallet.List(ctx)
		s.Stop()
		checkErr(err)

		data := make([][]string, len(addrs))
		for i, addr := range addrs {
			data[i] = []string{addr}
		}

		RenderTable(os.Stdout, []string{"address"}, data)
	},
}
