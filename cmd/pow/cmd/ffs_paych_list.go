package cmd

import (
	"context"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsPaychListCmd.Flags().StringP("token", "t", "", "token of the request")

	ffsPaychCmd.AddCommand(ffsPaychListCmd)
}

var ffsPaychListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the payment channels for the ffs instance",
	Long:  `List the payment channels for the ffs instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Retrieving payment channels...")
		s.Start()
		infos, err := fcClient.FFS.ListPayChannels(authCtx(ctx))
		checkErr(err)
		s.Stop()

		data := make([][]string, len(infos))
		for i, info := range infos {
			data[i] = []string{info.CtlAddr, info.Addr, ffs.PaychDirStr[info.Direction]}
		}
		Message("Payment channels:")
		RenderTable(os.Stdout, []string{"Ctrl Address", "Address", "Direction"}, data)
	},
}
