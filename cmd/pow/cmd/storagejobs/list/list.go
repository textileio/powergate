package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringP("cid", "c", "", "return results only for the specified cid")
	Cmd.Flags().Uint64P("limit", "l", 0, "limit the number of results returned")
	Cmd.Flags().BoolP("ascending", "a", false, "sort results ascending by time")
	Cmd.Flags().StringP("select", "s", "all", "return only results using the specified selector: all, queued, executing, final")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List storage jobs according to query flag options.",
	Long:  `List storage jobs according to query flag options.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		var sel client.ListSelect
		selIn := viper.GetString("select")
		switch selIn {
		case "all":
			sel = client.All
		case "queued":
			sel = client.Queued
		case "executing":
			sel = client.Executing
		case "final":
			sel = client.Final
		default:
			c.CheckErr(fmt.Errorf("invalid option for --select: %s", selIn))
		}

		conf := client.ListConfig{
			CidFilter: viper.GetString("cid"),
			Limit:     viper.GetUint64("limit"),
			Ascending: viper.GetBool("ascending"),
			Select:    sel,
		}

		res, err := c.PowClient.StorageJobs.List(c.MustAuthCtx(ctx), conf)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
