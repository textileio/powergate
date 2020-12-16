package verify

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "verify [addr] [hex-encoded-message] [hex-encoded-signature]",
	Short: "Verifies the signature of a message signed with a user wallet address.",
	Long:  "Verifies the signature of a message signed with a user wallet address.",
	Args:  cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		mb, err := hex.DecodeString(args[1])
		c.CheckErr(err)
		sb, err := hex.DecodeString(args[2])
		c.CheckErr(err)

		res, err := c.PowClient.Wallet.VerifyMessage(c.MustAuthCtx(ctx), args[0], mb, sb)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
