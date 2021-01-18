package send

import (
	"context"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
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
