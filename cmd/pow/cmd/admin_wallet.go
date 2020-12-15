package cmd

import (
	"context"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
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
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		format := viper.GetString("format")

		res, err := c.PowClient.Admin.Wallet.NewAddress(c.AdminAuthCtx(ctx), format)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

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
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.Wallet.Addresses(c.AdminAuthCtx(ctx))
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

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
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		amount, ok := new(big.Int).SetString(args[2], 10)
		if !ok {
			c.CheckErr(fmt.Errorf("parsing amount %v", args[2]))
		}

		_, err := c.PowClient.Admin.Wallet.SendFil(c.AdminAuthCtx(ctx), args[0], args[1], amount)
		c.CheckErr(err)
	},
}
