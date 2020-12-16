package config

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/config/apply"
	"github.com/textileio/powergate/cmd/pow/cmd/config/getdefault"
	"github.com/textileio/powergate/cmd/pow/cmd/config/remove"
	"github.com/textileio/powergate/cmd/pow/cmd/config/setdefault"
)

func init() {
	Cmd.AddCommand(apply.Cmd, getdefault.Cmd, remove.Cmd, setdefault.Cmd)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Provides commands to interact with cid storage configs",
	Long:  `Provides commands to interact with cid storage configs`,
}
