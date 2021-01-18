package admin

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/data"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/storageinfo"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/storagejobs"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/users"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/wallet"
)

func init() {
	Cmd.PersistentFlags().String("admin-token", "", "admin auth token")

	Cmd.AddCommand(
		data.Cmd,
		storagejobs.Cmd,
		storageinfo.Cmd,
		users.Cmd,
		wallet.Cmd,
	)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "admin",
	Short: "Provides admin commands",
	Long:  `Provides admin commands`,
}
