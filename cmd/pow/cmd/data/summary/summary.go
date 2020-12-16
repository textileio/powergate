package summary

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().BoolP("json", "j", false, "output data in raw json instead of an interactive ui")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "summary [optional cid1,cid2,...]",
	Short: "Get a summary about the current storage and jobs state of cids",
	Long:  `Get a summary about the current storage and jobs state of cids`,
	Args:  cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		json := viper.GetBool("json")

		if json {
			res, err := getSummary(cids)
			c.CheckErr(err)

			json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
			c.CheckErr(err)

			fmt.Println(string(json))
			return
		}

		p := tea.NewProgram(model{cids: cids})
		if err := p.Start(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func getSummary(cids []string) (*userPb.CidSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
	defer cancel()
	return c.PowClient.Data.CidSummary(c.MustAuthCtx(ctx), cids...)
}
