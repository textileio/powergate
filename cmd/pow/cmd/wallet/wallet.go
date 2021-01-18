package wallet

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/addrs"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/balance"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/newaddr"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/send"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/sign"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet/verify"
)

func init() {
	Cmd.AddCommand(addrs.Cmd, balance.Cmd, newaddr.Cmd, send.Cmd, sign.Cmd, verify.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "wallet",
	Short: "Provides commands about filecoin wallets",
	Long:  `Provides commands about filecoin wallets`,
}
