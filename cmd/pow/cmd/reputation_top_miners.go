package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	topMinersCmd.Flags().Int32P("limit", "l", -1, "limit the number of results")

	reputationCmd.AddCommand(topMinersCmd)
}

var topMinersCmd = &cobra.Command{
	Use:   "topMiners",
	Short: "Fetches a list of the currently top rated miners",
	Long:  `Fetches a list of the currently top rated miners`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		limit, err := cmd.Flags().GetInt32("limit")
		checkErr(err)

		res, err := fcClient.Reputation.GetTopMiners(ctx, limit)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
