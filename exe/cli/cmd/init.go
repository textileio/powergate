package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	initCmd.Flags().StringP("address", "a", "", "the default wallet address")
	initCmd.Flags().Uint64P("duration", "d", 0, "the duration to store the file for")
	initCmd.Flags().Uint64P("maxPrice", "m", 0, "max price for asks query")
	initCmd.Flags().Uint64P("pieceSize", "p", 0, "piece size for asks query")
	initCmd.Flags().StringSliceP("wallets", "w", []string{}, "existing wallets to switch between")

	viper.BindPFlag("address", initCmd.Flags().Lookup("address"))
	viper.BindPFlag("duration", initCmd.Flags().Lookup("duration"))
	viper.BindPFlag("maxPrice", initCmd.Flags().Lookup("maxPrice"))
	viper.BindPFlag("pieceSize", initCmd.Flags().Lookup("pieceSize"))
	viper.BindPFlag("wallets", initCmd.Flags().Lookup("wallets"))

	viper.SetDefault("address", "")
	viper.SetDefault("duration", 100)
	viper.SetDefault("maxPrice", 200)
	viper.SetDefault("pieceSize", 1024)
	viper.SetDefault("wallets", []string{})

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
