package storagejobs

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/storagejobs/list"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/storagejobs/summary"
)

func init() {
	Cmd.AddCommand(summary.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:     "storage-jobs",
	Aliases: []string{"storage-job"},
	Short:   "Provides admin jobs commands",
	Long:    `Provides admin jobs commands`,
}
