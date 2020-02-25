package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpaInfoCmd.Flags().StringP("token", "t", "", "token of the request")

	fpaCmd.AddCommand(fpaInfoCmd)
}

var fpaInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info from fpa instance",
	Long:  `Get info from fpa instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		token := viper.GetString("token")

		if token == "" {
			Fatal(errors.New("get requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpatoken"), token)

		s := spin.New("%s Retrieving instance info...")
		s.Start()
		info, err := fcClient.Fpa.Info(ctx)
		checkErr(err)
		s.Stop()
		Message("Information from instance ID %s:", aurora.White(info.Id).Bold())
		Message("Wallet %s has balance %d", aurora.White(info.Wallet.Address), aurora.Green(info.Wallet.Balance))

		Message("Pinned cids:")
		data := make([][]string, len(info.Pins))
		for i, cid := range info.Pins {
			data[i] = []string{cid}
		}
		RenderTable(os.Stdout, []string{"cid"}, data)
	},
}
