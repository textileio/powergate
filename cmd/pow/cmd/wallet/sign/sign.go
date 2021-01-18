package sign

import (
	"context"
	"encoding/hex"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "sign [hex-encoded-message]",
	Short: "Signs a message with user wallet addresses.",
	Long:  "Signs a message using all wallet addresses associated with the user",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		b, err := hex.DecodeString(args[0])
		c.CheckErr(err)

		res, err := c.PowClient.Wallet.Addresses(c.MustAuthCtx(ctx))
		c.CheckErr(err)

		data := make([][]string, len(res.Addresses))
		for i, a := range res.Addresses {
			signRes, err := c.PowClient.Wallet.SignMessage(c.MustAuthCtx(ctx), a.Address, b)
			c.CheckErr(err)
			data[i] = []string{a.Address, hex.EncodeToString(signRes.Signature)}
		}

		c.RenderTable(os.Stdout, []string{"address", "signature"}, data)
	},
}
