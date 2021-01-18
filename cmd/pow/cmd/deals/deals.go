package deals

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/deals/retrievals"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/deals/storage"
)

func init() {
	Cmd.AddCommand(retrievals.Cmd, storage.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "deals",
	Short: "Provides commands to view Filecoin deal information",
	Long:  `Provides commands to view Filecoin deal information`,
}
