package new

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringP("format", "f", "bls", "Optionally specify address format bls or secp256k1")
}

// Cmd is the command.
var Cmd = &cobra.Command{
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
