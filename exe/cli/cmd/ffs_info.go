package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsInfoCmd.Flags().StringP("token", "t", "", "token of the request")

	ffsCmd.AddCommand(ffsInfoCmd)
}

var ffsInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info from ffs instance",
	Long:  `Get info from ffs instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Retrieving instance info...")
		s.Start()
		resp, err := fcClient.Ffs.Info(authCtx(ctx))
		checkErr(err)
		s.Stop()

		Message("FFS instance id: %v", resp.Info.ID)

		data := make([][]string, len(resp.Info.Balances))
		for i, balance := range resp.Info.Balances {
			isDefault := ""
			if balance.Addr.Addr == resp.Info.DefaultConfig.Cold.Filecoin.Addr {
				isDefault = "yes"
			}
			data[i] = []string{balance.Addr.Name, balance.Addr.Addr, fmt.Sprintf("%v", balance.Balance), isDefault}
		}
		Message("Wallet addresses:")
		RenderTable(os.Stdout, []string{"name", "address", "balance", "default"}, data)

		bytes, err := json.MarshalIndent(resp.Info.DefaultConfig, "", "  ")
		checkErr(err)

		Message("Default storage config:\n %v", string(bytes))

		Message("Pinned cids:")
		data = make([][]string, len(resp.Info.Pins))
		for i, cid := range resp.Info.Pins {
			data[i] = []string{cid}
		}
		RenderTable(os.Stdout, []string{"cid"}, data)
	},
}
