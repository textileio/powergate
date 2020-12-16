package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
	client "github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/cmd/pow/cmd/admin"
	"github.com/textileio/powergate/cmd/pow/cmd/config"
	"github.com/textileio/powergate/cmd/pow/cmd/data"
	"github.com/textileio/powergate/cmd/pow/cmd/deals"
	"github.com/textileio/powergate/cmd/pow/cmd/id"
	"github.com/textileio/powergate/cmd/pow/cmd/storagejobs"
	"github.com/textileio/powergate/cmd/pow/cmd/version"
	"github.com/textileio/powergate/cmd/pow/cmd/wallet"
	c "github.com/textileio/powergate/cmd/pow/common"
)

func init() {
	cobra.OnInitialize(initConfig)
	Cmd.Flags().BoolP("version", "v", false, "display version information for pow and the connected server")
	Cmd.PersistentFlags().String("serverAddress", "127.0.0.1:5002", "address of the powergate service api")
	Cmd.PersistentFlags().StringP("token", "t", "", "user auth token")

	Cmd.AddCommand(admin.Cmd, config.Cmd, data.Cmd, deals.Cmd, id.Cmd, storagejobs.Cmd, version.Cmd, wallet.Cmd, docsCmd)
}

func initConfig() {
	viper.SetEnvPrefix("POW")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:               "pow",
	Short:             "A client for storage and retreival of powergate data",
	Long:              `A client for storage and retreival of powergate data`,
	DisableAutoGenTag: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlag("serverAddress", cmd.Root().PersistentFlags().Lookup("serverAddress"))
		c.CheckErr(err)

		target := viper.GetString("serverAddress")

		c.PowClient, err = client.NewClient(target)
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		v, err := cmd.Flags().GetBool("version")
		c.CheckErr(err)
		if v {
			version.Cmd.Run(cmd, args)
		} else {
			err := cmd.Help()
			c.CheckErr(err)
		}
	},
}

var docsCmd = &cobra.Command{
	Use:    "docs [outdir]",
	Short:  "Generate markdown docs for pow command",
	Long:   `Generate markdown docs for pow command`,
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			c.CheckErr(err)
		}
		err := doc.GenMarkdownTree(Cmd, args[0])
		c.CheckErr(err)
	},
}
