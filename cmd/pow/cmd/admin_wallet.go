package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	adminWalletNewCmd.Flags().StringP("format", "f", "bls", "Optionally specify address format bls or secp256k1")

	adminWalletCmd.AddCommand(
		adminWalletNewCmd,
		adminWalletAddrsCmd,
		adminWalletSendCmd,
	)
}

var adminWalletNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new walllet address.",
	Long:  `Creates a new wallet address.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		format := viper.GetString("format")

		res, err := powClient.Admin.Wallet.NewAddress(adminAuthCtx(ctx), format)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminWalletAddrsCmd = &cobra.Command{
	Use:   "addrs",
	Short: "List all addresses associated with this Powergate.",
	Long:  `List all addresses associated with this Powergate.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.Admin.Wallet.ListAddresses(adminAuthCtx(ctx))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminWalletSendCmd = &cobra.Command{
	Use:   "send [from] [to] [amount]",
	Short: "Sends FIL from an address associated with this Powergate to any other address.",
	Long:  `Sends FIL from an address associated with this Powergate to any other address.`,
	Args:  cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		amount, err := strconv.ParseInt(args[2], 10, 64)
		checkErr(err)

		_, err = powClient.Admin.Wallet.SendFil(adminAuthCtx(ctx), args[0], args[1], amount)
		checkErr(err)
	},
}
