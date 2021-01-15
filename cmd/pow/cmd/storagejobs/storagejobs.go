package storagejobs

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/cancel"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/cancelexecuting"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/cancelqueued"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/get"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/list"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/storageconfig"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/summary"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs/watch"
)

func init() {
	Cmd.AddCommand(
		cancel.Cmd,
		cancelexecuting.Cmd,
		cancelqueued.Cmd,
		get.Cmd,
		list.Cmd,
		storageconfig.Cmd,
		summary.Cmd,
		watch.Cmd,
	)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:     "storage-jobs",
	Aliases: []string{"storage-job"},
	Short:   "Provides commands to query for storage jobs in various states",
	Long:    `Provides commands to query for storage jobs in various statess`,
}
