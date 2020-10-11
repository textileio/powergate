package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		s := spin.New(fmt.Sprintf("%s Signing message with addresses...", "%s"))
		s.Start()
		res, err := fcClient.FFS.Addrs(authCtx(ctx))
		checkErr(err)
		data := make([][]string, len(res.Addrs))
		for _, a := range res.Addrs {
			sig, err := fcClient.FFS.SignMessage(authCtx(ctx), a.Addr, b)
			checkErr(err)
			data = append(data, []string{a.Addr, hex.EncodeToString(sig)})
		}
		s.Stop()

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

		s := spin.New(fmt.Sprintf("%s Verifying signature...", "%s"))
		s.Start()
		ok, err := fcClient.FFS.VerifyMessage(authCtx(ctx), args[0], mb, sb)
		s.Stop()
		checkErr(err)
		if ok {
			_, err := emoji.Println(":heavy_check_mark: The signature corresponds to the wallet address.")
			checkErr(err)
		} else {
			_, err := emoji.Println(":x: The signature doesn't correspond to the wallet address.")
			checkErr(err)
		}
	},
}
