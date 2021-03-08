package addrs

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	Cmd.Flags().BoolP("detailed", "d", false, "include extra information about the wallet-address")
}

// Cmd is the command.
var Cmd = &cobra.Command{
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

		detailed, err := cmd.Flags().GetBool("detailed")
		c.CheckErr(err)

		var res proto.Message
		if detailed {
			res, err = c.PowClient.Admin.Wallet.AddressesDetailed(c.AdminAuthCtx(ctx))
		} else {
			res, err = c.PowClient.Admin.Wallet.Addresses(c.AdminAuthCtx(ctx))
		}
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
