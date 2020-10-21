package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	faultsCmd.AddCommand(getFaultsCmd)
}

var getFaultsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the faults index",
	Long:  `Get the faults index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Faults.Get(ctx)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.Index)
		checkErr(err)

		fmt.Println(string(json))
	},
}
