package cmd

import (
	"context"
	"fmt"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	minersCmd.AddCommand(getMinersCmd)
}

var getMinersCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the miners index",
	Long:  `Get the miners index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Miners.Get(ctx)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))

		Message("%v", aurora.Blue("Miner metadata:").Bold())
		cmd.Println()
	},
}
