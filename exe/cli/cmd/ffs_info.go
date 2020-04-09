package cmd

import (
	"context"
	"os"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
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
		Message("Information from instance ID %s:", aurora.White(resp.Info.ID).Bold())
		Message("Wallet %s has balance %d", aurora.White(resp.Info.Wallet.Address), aurora.Green(resp.Info.Wallet.Balance))

		Message("Pinned cids:")
		data := make([][]string, len(resp.Info.Pins))
		for i, cid := range resp.Info.Pins {
			data[i] = []string{cid}
		}
		RenderTable(os.Stdout, []string{"cid"}, data)
	},
}
