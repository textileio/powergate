package jobs

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/list"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/summary"
)

func init() {
	Cmd.AddCommand(summary.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"job"},
	Short:   "Provides admin jobs commands",
	Long:    `Provides admin jobs commands`,
}
