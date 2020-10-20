package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	ffsCmd.AddCommand(signCmd)
	ffsCmd.AddCommand(verifyCmd)
}

var signCmd = &cobra.Command{
	Use:   "sign [hex-encoded-message]",
	Short: "Signs a message with FFS wallet addresses.",
	Long:  "Signs a message using all wallet addresses associated with the instance",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		b, err := hex.DecodeString(args[0])
		checkErr(err)

		res, err := fcClient.FFS.Addrs(mustAuthCtx(ctx))
		checkErr(err)

		data := make([][]string, len(res.Addrs))
		for i, a := range res.Addrs {
			signRes, err := fcClient.FFS.SignMessage(mustAuthCtx(ctx), a.Addr, b)
			checkErr(err)
			data[i] = []string{a.Addr, hex.EncodeToString(signRes.Signature)}
		}

		RenderTable(os.Stdout, []string{"address", "signature"}, data)
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify [addr] [hex-encoded-message] [hex-encoded-signature]",
	Short: "Verifies the signature of a message signed with a FFS wallet address.",
	Long:  "Verifies the signature of a message signed with a FFS wallet address.",
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

		res, err := fcClient.FFS.VerifyMessage(mustAuthCtx(ctx), args[0], mb, sb)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
