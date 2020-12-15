package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun         bool
	mainnet        bool
	resume         bool
	maxStagedBytes int64
	maxDealBytes   int64
	minDealBytes   int64
	hiddenFiles    bool
	taskFolder     string
	resultsOut     string
	concurrent     int64
	ipfsrevproxy   string
	retryErrors    bool
)

func init() {
	runCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Run through steps without pushing data to Powergate API.")
	runCmd.Flags().BoolVarP(&mainnet, "mainnet", "m", false, "Sets staging limits based on localnet or mainnet.")
	runCmd.Flags().BoolVarP(&retryErrors, "retry-errors", "e", false, "Retry tasks with error status")
	runCmd.Flags().StringVar(&taskFolder, "folder", "f", "Folder with organized tasks of directories or files")
	runCmd.Flags().BoolVarP(&hiddenFiles, "include-all", "i", false, "Include hidden files & folders from top level folder")
	runCmd.Flags().StringVar(&ipfsrevproxy, "ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
	runCmd.Flags().StringVarP(&resultsOut, "output", "o", "results.json", "The location to store intermediate and final results.")
	runCmd.Flags().BoolVarP(&resume, "resume", "r", false, "Resume tasks stored in local results output file.")

	// Not included in the public commands
	runCmd.Flags().Int64Var(&concurrent, "concurrent", 1000, "Max concurrent tasks being processed")
	runCmd.Flags().Int64Var(&maxStagedBytes, "max-staged-bytes", 0, "Maximum bytes of all tasks queued on staging")
	runCmd.Flags().Int64Var(&maxDealBytes, "max-deal-bytes", 0, "Maximum bytes of a single deal")
	runCmd.Flags().Int64Var(&minDealBytes, "min-deal-bytes", 0, "Minimum bytes of a single deal")
	err := runCmd.Flags().MarkHidden("concurrent")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("max-staged-bytes")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("max-deal-bytes")
	checkErr(err)
	err = runCmd.Flags().MarkHidden("min-deal-bytes")
	checkErr(err)

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a storage task pipeline",
	Long:  `Run a storage task pipeline to migrate large collections of data to Filecoin`,
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
				minDealBytes = 9000
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if maxStagedBytes < maxDealBytes {
			Fatal(fmt.Errorf("max deal size (%d) is larger than max staging size (%d)", maxDealBytes, maxStagedBytes))
		}

		allTasks := make([]Task, 0)
		if !resume {
			// Read all tasks from the input folder
			unfiltered, err := pathToTasks(taskFolder)
			checkErr(err)
			Message("Input tasks: %d", len(unfiltered))
			allTasks = filterTasks(unfiltered)
		}

		// Ensure there aren't existing tasks in the same queue
		knownTasks, err := openResults(resultsOut)
		checkErr(err)
		Message("Existing tasks: %d", len(knownTasks))
		allTasks = mergeNewTasks(knownTasks, allTasks)

		pendingTasks := cleanErrors(removeComplete(allTasks), retryErrors)

		if len(pendingTasks) == 0 {
			return
		}

		Message("Starting tasks: %d", len(pendingTasks))
		time.Sleep(1 * time.Second)

		// Init or update our intermediate results
		err = storeResults(resultsOut, allTasks)
		checkErr(err)

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
			mainnet:         mainnet,
		}
		storageConfig, err := getStorageConfig(rc)
		checkErr(err)
		storageConfig.Hot.Enabled = false
		rc.storageConfig = storageConfig

		OutputProgress(pendingTasks, allTasks, rc)
	},
}

// OutputProgress starts the pipeline and outputs progress to file or terminal.
func OutputProgress(pendingTasks []Task, allTasks []Task, conf PipelineConfig) {
	progress := uiprogress.New()
	progress.Start()
	progressBars := make(map[string](chan Task))

	taskUpdates := Start(pendingTasks, conf)

	j := 1
	var wg sync.WaitGroup
	for task := range taskUpdates {
		allTasks, change := updateTasks(allTasks, task)
		if change {
			if _, found := progressBars[task.Name]; !found {
				progressBars[task.Name] = make(chan Task)
				wg.Add(1)
				go progressBar(progress, progressBars[task.Name], j, len(pendingTasks), &wg)
				j++
			}
			if task.Stage > DryRunComplete {
				progressBars[task.Name] <- task
			}
			err := storeResults(resultsOut, allTasks)
			checkErr(err)
		} else {
			Message(fmt.Sprintf("%s %d", task.Name, task.Stage))
		}
	}

	go func() {
		wg.Wait()
		for _, c := range progressBars {
			close(c)
		}
	}()

	progress.Stop()
}
