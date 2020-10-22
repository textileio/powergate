package cmd

import (
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
		// ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		// defer cancel()

		// res, err := fcClient.Wallet.List(ctx)
		// checkErr(err)

		// json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		// checkErr(err)

		// fmt.Println(string(json))
	},
}
