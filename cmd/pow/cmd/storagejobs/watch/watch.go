package watch

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "watch [jobid,...]",
	Short: "Watch for storage job status updates",
	Long:  `Watch for storage job status updates`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jobIds := strings.Split(args[0], ",")
		c.WatchJobIds(jobIds...)
	},
}
