package cmd

import (
	"context"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	importCmd.Flags().BoolP("car", "c", false, "Specifies if the data is already CAR encoded")

	dealsCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import [file path]",
	Short: "Import data into the Filecoin client",
	Long:  `Import data into the Filecoin client in preparation for making a storage deal using the store command`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()

		path := args[0]
		isCAR := viper.GetBool("car")

		file, err := os.Open(path)
		checkErr(err)
		defer func() { checkErr(file.Close()) }()

		s := spin.New("%s Importing file...")
		s.Start()
		cid, size, err := fcClient.Deals.Import(ctx, file, isCAR)
		s.Stop()
		checkErr(err)

		Success("Imported %v bytes resulting in data cid %v", size, cid.String())
	},
}
