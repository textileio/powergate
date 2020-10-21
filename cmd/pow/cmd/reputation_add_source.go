package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	addSourceCmd.Flags().StringP("id", "i", "", "id of the miner to add")
	addSourceCmd.Flags().StringP("address", "a", "", "multiaddress of the miner to add")

	reputationCmd.AddCommand(addSourceCmd)
}

var addSourceCmd = &cobra.Command{
	Use:   "addSource",
	Short: "Adds a new external source to be considered for reputation generation",
	Long:  `Aadds a new external source to be considered for reputation generation`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		id := cmd.Flag("id").Value.String()
		address := cmd.Flag("address").Value.String()

		if id == "" {
			Fatal(errors.New("must provide a miner id"))
		}

		if address == "" {
			Fatal(errors.New("must provide a miner address"))
		}

		_, err := fcClient.Reputation.AddSource(ctx, id, address)
		checkErr(err)
	},
}
