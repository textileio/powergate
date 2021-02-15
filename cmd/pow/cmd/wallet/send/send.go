package send

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "send [from address] [to address] [amount]",
	Short: "Send fil from one managed address to any other address",
	Long:  `Send fil from one managed address to any other address`,
	Args:  cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		from := args[0]
		to := args[1]

		amount, ok := new(big.Int).SetString(args[2], 10)
		if !ok {
			c.CheckErr(fmt.Errorf("parsing amount %v", args[2]))
		}

		res, err := c.PowClient.Wallet.SendFil(c.MustAuthCtx(ctx), from, to, amount)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
