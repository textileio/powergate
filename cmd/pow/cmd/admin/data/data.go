package data

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/data/gcstaged"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/data/pinnedcids"
)

func init() {
	Cmd.AddCommand(gcstaged.Cmd, pinnedcids.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "data",
	Short: "Provides admin data commands",
	Long:  `Provides admin data commands`,
}
