package cmd

import (
	"context"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsWalletAddrCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsWalletAddrCmd)
}

var ffsWalletAddrCmd = &cobra.Command{
	Use:   "waddr",
	Short: "Returns the FFS instance wallet address",
	Long:  `Returns the FFS instance wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		s := spin.New("%s Getting FFS instance wallet address...")
		s.Start()
		addr, err := fcClient.Ffs.WalletAddr(authCtx(ctx))
		s.Stop()
		checkErr(err)
		Message("FFS instance wallet address: %s", addr)
	},
}
