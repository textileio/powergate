package newaddr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringP("format", "f", "", "Optionally specify address format bls or secp256k1")
	Cmd.Flags().BoolP("default", "d", false, "Make the new address the user default")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "new-addr [name]",
	Short: "Create a new wallet address",
	Long:  `Create a new wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			c.Fatal(errors.New("must provide a name for the address"))
		}

		format := viper.GetString("format")
		makeDefault := viper.GetBool("default")

		var opts []client.NewAddressOption
		if format != "" {
			opts = append(opts, client.WithAddressType(format))
		}
		if makeDefault {
			opts = append(opts, client.WithMakeDefault(makeDefault))
		}

		res, err := c.PowClient.Wallet.NewAddress(c.MustAuthCtx(ctx), args[0], opts...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
