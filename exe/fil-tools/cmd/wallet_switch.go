package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	switchCmd.Flags().StringP("wallet", "w", "", "wallet address to make the new default")

	walletCmd.AddCommand(switchCmd)
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the default wallet address",
	Long:  `Switch the default wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("wallets", []string{})
	},
	Run: func(cmd *cobra.Command, args []string) {
		wallet, err := cmd.Flags().GetString("wallet")
		checkErr(err)

		wallets := viper.GetStringSlice("wallets")

		if len(wallet) > 0 {
			viper.Set("address", wallet)
			if !contains(wallet, wallets) {
				wallets = append(wallets, wallet)
				viper.Set("wallets", wallets)
			}
		} else {
			prompt := promptui.Select{
				Label: "Select a default wallet address",
				Items: wallets,
			}
			index, _, err := prompt.Run()
			checkErr(err)
			viper.Set("address", wallets[index])
		}

		err = viper.WriteConfig()
		checkErr(err)

		Success("Default wallet address set to %v", viper.GetString("address"))
	},
}

func contains(target string, values []string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
