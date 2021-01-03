package jobs

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/executing"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/latestfinal"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/latestsuccessful"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/list"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/queued"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/summary"
)

func init() {
	Cmd.AddCommand(executing.Cmd, latestfinal.Cmd, latestsuccessful.Cmd, queued.Cmd, summary.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"job"},
	Short:   "Provides admin jobs commands",
	Long:    `Provides admin jobs commands`,
}
