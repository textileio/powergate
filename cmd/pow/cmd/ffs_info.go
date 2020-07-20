package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
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
		info, err := fcClient.FFS.Info(authCtx(ctx))
		checkErr(err)
		s.Stop()

		Message("FFS instance id: %v", info.ID.String())

		data := make([][]string, len(info.Balances))
		for i, balance := range info.Balances {
			isDefault := ""
			if balance.Addr == info.DefaultConfig.Cold.Filecoin.Addr {
				isDefault = "yes"
			}
			data[i] = []string{balance.Name, balance.Addr, balance.Type, fmt.Sprintf("%v", balance.Balance), isDefault}
		}
		Message("Wallet addresses:")
		RenderTable(os.Stdout, []string{"name", "address", "type", "balance", "default"}, data)

		bytes, err := json.MarshalIndent(info.DefaultConfig, "", "  ")
		checkErr(err)

		Message("Default storage config:\n%v", string(bytes))

		Message("Pinned cids:")
		data = make([][]string, len(info.Pins))
		for i, cid := range info.Pins {
			data[i] = []string{util.CidToString(cid)}
		}
		RenderTable(os.Stdout, []string{"cid"}, data)
	},
}
