package cmd

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	walletCmd.AddCommand(walletVerifyCmd)
}

var walletVerifyCmd = &cobra.Command{
	Use:   "verify [addr] [hex-encoded-message] [hex-encoded-signature]",
	Short: "Verifies the signature of a message signed with a storage profile wallet address.",
	Long:  "Verifies the signature of a message signed with a storage profile wallet address.",
	Args:  cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		mb, err := hex.DecodeString(args[1])
		checkErr(err)
		sb, err := hex.DecodeString(args[2])
		checkErr(err)

		res, err := powClient.Wallet.VerifyMessage(mustAuthCtx(ctx), args[0], mb, sb)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
