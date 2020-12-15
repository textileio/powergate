package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	resumeCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Run through steps without pushing data to Powergate API.")
	resumeCmd.Flags().BoolVarP(&mainnet, "mainnet", "n", false, "Sets staging limits based on localnet or mainnet.")
	resumeCmd.Flags().BoolVarP(&retryErrors, "retry", "e", false, "Retry tasks with error status")
	resumeCmd.Flags().StringVar(&ipfsrevproxy, "ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
	resumeCmd.Flags().StringVarP(&resultsOut, "results", "r", "results.json", "The of ongoing tasks to resume.")

	// Not included in the public commands
	resumeCmd.Flags().Int64Var(&concurrent, "concurrent", 1000, "Max concurrent tasks being processed")
	resumeCmd.Flags().Int64Var(&maxStagedBytes, "max-staged-bytes", 0, "Maximum bytes of all tasks queued on staging")
	resumeCmd.Flags().Int64Var(&maxDealBytes, "max-deal-bytes", 0, "Maximum bytes of a single deal")
	resumeCmd.Flags().Int64Var(&minDealBytes, "min-deal-bytes", 0, "Minimum bytes of a single deal")

	err := runCmd.Flags().MarkHidden("concurrent")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("max-staged-bytes")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("max-deal-bytes")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("min-deal-bytes")
	checkErr(err)

	rootCmd.AddCommand(resumeCmd)
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a storage task pipeline",
	Long:  `Resume a storage task pipeline to complete migration to Filecoin`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
		// set default variables
		if mainnet {
			if maxStagedBytes == 0 {
				maxStagedBytes = 26843545600 // 25Gib
			}
			if maxDealBytes == 0 {
				maxDealBytes = 4294967296 // 4Gib
			}
			if minDealBytes == 0 {
				minDealBytes = 67108864 // 0.5Gib
			}
		} else { // localnet
			if maxStagedBytes == 0 {
				maxStagedBytes = 524288
			}
			if maxDealBytes == 0 {
				maxDealBytes = 524288
			}
			if minDealBytes == 0 {
				minDealBytes = 5120
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if maxStagedBytes < maxDealBytes {
			Fatal(fmt.Errorf("Max deal size (%d) is larger than max staging size (%d)", maxDealBytes, maxStagedBytes))
		}

		// Ensure there aren't existing tasks in the same queue
		knownTasks, err := openResults(resultsOut)
		checkErr(err)

		pendingTasks := cleanErrors(removeComplete(knownTasks), retryErrors)

		Message("Total pending tasks: %d", len(pendingTasks))
		time.Sleep(1 * time.Second)

		if len(pendingTasks) == 0 {
			return
		}

		// Determine the max concurrent we need vs allowed
		if int64(len(pendingTasks)) < concurrent {
			concurrent = int64(len(pendingTasks))
		}

		rc := PipelineConfig{
			token:           viper.GetString("token"),
			serverAddress:   viper.GetString("serverAddress"),
			ipfsrevproxy:    ipfsrevproxy,
			maxStagingBytes: maxStagedBytes,
			minDealBytes:    minDealBytes,
			concurrent:      concurrent,
			dryRun:          dryRun,
		}

		storageConfig, err := getStorageConfig(rc)
		checkErr(err)
		storageConfig.Hot.Enabled = false
		rc.storageConfig = storageConfig

		trackProgress(pendingTasks, knownTasks, rc)
	},
}
