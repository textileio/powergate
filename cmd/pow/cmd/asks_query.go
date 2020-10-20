package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/index/ask/rpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	queryCmd.Flags().Uint64P("maxPrice", "m", 0, "max price of the asks to query")
	queryCmd.Flags().Uint64P("pieceSize", "p", 0, "piece size of the asks to query")
	queryCmd.Flags().Int32P("limit", "l", -1, "limit the number of results")
	queryCmd.Flags().Int32P("offset", "o", -1, "offset of results")

	asksCmd.AddCommand(queryCmd)
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query the available asks",
	Long:  `Query the available asks`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		mp := viper.GetUint64("maxPrice")
		ps := viper.GetUint64("pieceSize")
		l := viper.GetInt32("limit")
		o := viper.GetInt32("offset")

		if mp == 0 {
			Fatal(errors.New("maxPrice must be > 0"))
		}

		if ps == 0 {
			Fatal(errors.New("pieceSize must be > 0"))
		}

		q := &rpc.Query{
			MaxPrice:  mp,
			PieceSize: ps,
			Limit:     l,
			Offset:    o,
		}

		res, err := fcClient.Asks.Query(ctx, q)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
