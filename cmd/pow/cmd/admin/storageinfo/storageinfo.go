package storageinfo

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/storageinfo/get"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/storageinfo/list"
)

func init() {
	Cmd.AddCommand(get.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "storage-info",
	Short: "Provides admin storage info commands",
	Long:  `Provides admin storage info commands`,
}
