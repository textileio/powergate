package wallet

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/wallet/addrs"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/wallet/new"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/wallet/send"
)

func init() {
	Cmd.AddCommand(addrs.Cmd, new.Cmd, send.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "wallet",
	Short: "Provides admin wallet commands",
	Long:  `Provides admin wallet commands`,
}
