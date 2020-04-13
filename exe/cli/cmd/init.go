package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	initCmd.Flags().StringP("address", "a", "<a-wallet-address>", "Sets the default wallet address for the config file")
	initCmd.Flags().IntP("duration", "d", 1000000, "Sets the default duration for the config file")
	initCmd.Flags().IntP("maxPrice", "m", 100, "Sets the default maxPrice for the config file")
	initCmd.Flags().IntP("pieceSize", "p", 1024, "Sets the default pieceSize for the config file")
	initCmd.Flags().StringSliceP("wallets", "w", []string{}, "Sets the wallet address list for the config file")

	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a config file with the provided values or defaults",
	Long:  `Initializes a config file with the provided values or defaults`,
	PreRun: func(cmd *cobra.Command, args []string) {
		checkErr(viper.BindPFlag("address", cmd.Flags().Lookup("address")))
		checkErr(viper.BindPFlag("duration", cmd.Flags().Lookup("duration")))
		checkErr(viper.BindPFlag("maxPrice", cmd.Flags().Lookup("maxPrice")))
		checkErr(viper.BindPFlag("pieceSize", cmd.Flags().Lookup("pieceSize")))
		checkErr(viper.BindPFlag("wallets", cmd.Flags().Lookup("wallets")))
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigType("yaml")
		err := viper.SafeWriteConfig()
		checkErr(err)
	},
}
