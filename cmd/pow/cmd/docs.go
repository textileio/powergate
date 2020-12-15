package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	c "github.com/textileio/powergate/cmd/pow/common"
)

func init() {
	rootCmd.AddCommand(docsCmd)
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
		err := doc.GenMarkdownTree(rootCmd, args[0])
		c.CheckErr(err)
	},
}
