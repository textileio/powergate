package users

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/users/create"
	"github.com/textileio/powergate/v2/cmd/pow/cmd/admin/users/list"
)

func init() {
	Cmd.AddCommand(create.Cmd, list.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Provides admin users commands",
	Long:    `Provides admin users commands`,
}
