package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(storageJobsCmd)
}

var storageJobsCmd = &cobra.Command{
	Use:     "storage-jobs",
	Aliases: []string{"storage-job"},
	Short:   "Provides commands to query for storage jobs in various states",
	Long:    `Provides commands to query for storage jobs in various statess`,
}
