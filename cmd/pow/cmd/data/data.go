package data

import (
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/cmd/pow/cmd/data/get"
	"github.com/textileio/powergate/cmd/pow/cmd/data/info"
	"github.com/textileio/powergate/cmd/pow/cmd/data/log"
	"github.com/textileio/powergate/cmd/pow/cmd/data/replace"
	"github.com/textileio/powergate/cmd/pow/cmd/data/stage"
	"github.com/textileio/powergate/cmd/pow/cmd/data/summary"
)

func init() {
	Cmd.AddCommand(get.Cmd, info.Cmd, log.Cmd, replace.Cmd, stage.Cmd, summary.Cmd)
}

// Cmd provides commands to interact with general data APIs.
var Cmd = &cobra.Command{
	Use:   "data",
	Short: "Provides commands to interact with general data APIs",
	Long:  `Provides commands to interact with general data APIs`,
}
