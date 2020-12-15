package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/buildinfo"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information for fts and the connected server",
	Long:  `Display version information for fts and the connected server`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		Message("fts build info:")
		renderTable(
			os.Stdout,
			[]string{},
			[][]string{
				{"Version", buildinfo.Version},
				{"Git Summary", buildinfo.GitSummary},
				{"Git Branch", buildinfo.GitBranch},
				{"Git State", buildinfo.GitState},
				{"Git Commit", buildinfo.GitCommit},
				{"Build Date", buildinfo.BuildDate},
			},
		)

		s := spin.New("%s Getting Powergate server build info...")
		s.Start()
		info, err := powClient.BuildInfo(ctx)
		s.Stop()
		checkErr(err)

		fmt.Print("\n")

		Message("powergate server build info:")
		renderTable(
			os.Stdout,
			[]string{},
			[][]string{
				{"Version", info.Version},
				{"Git Summary", info.GitSummary},
				{"Git Branch", info.GitBranch},
				{"Git State", info.GitState},
				{"Git Commit", info.GitCommit},
				{"Build Date", info.BuildDate},
			},
		)
	},
}
