package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(storageInfoCmd)
}

var storageInfoCmd = &cobra.Command{
	Use:   "storage-info",
	Short: "Provides commands to get and query cid storage info.",
	Long:  `Provides commands to get and query cid storage info.`,
}
