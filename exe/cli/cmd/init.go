package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a config file with the provided values or defaults",
	Long:  `Initializes a config file with the provided values or defaults`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigType("yaml")
		err := viper.SafeWriteConfig()
		checkErr(err)
	},
}
