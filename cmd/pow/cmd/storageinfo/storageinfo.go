package storageinfo

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/storageinfo/get"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/storageinfo/list"
)

func init() {
	Cmd.AddCommand(get.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "storage-info",
	Short: "Provides commands to get and query cid storage info.",
	Long:  `Provides commands to get and query cid storage info.`,
}
