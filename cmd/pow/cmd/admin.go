package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.PersistentFlags().String("admin-token", "", "admin auth token")

	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(
		adminJobsCmd,
		adminUsersCmd,
		adminWalletCmd,
		adminDataCmd,
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

var adminUsersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user"},
	Short:   "Provides admin users commands",
	Long:    `Provides admin users commands`,
}

var adminWalletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Provides admin wallet commands",
	Long:  `Provides admin wallet commands`,
}

var adminDataCmd = &cobra.Command{
	Use:   "data",
	Short: "Provides admin data commands",
	Long:  `Provides admin data commands`,
}
