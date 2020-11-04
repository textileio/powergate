package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.PersistentFlags().String("admin-token", "", "admin auth token")

	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(
		adminJobsCmd,
		adminProfilesCmd,
		adminWalletCmd,
	)
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Provides admin commands",
	Long:  `Provides admin commands`,
}

var adminJobsCmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"job"},
	Short:   "Provides admin jobs commands",
	Long:    `Provides admin jobs commands`,
}

var adminProfilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile"},
	Short:   "Provides admin storage profile commands",
	Long:    `Provides admin storage profile commands`,
}

var adminWalletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Provides admin wallet commands",
	Long:  `Provides admin wallet commands`,
}
