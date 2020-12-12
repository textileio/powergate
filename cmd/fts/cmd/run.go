package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

var (
	dryRun         bool
	quiet          bool
	verbose        bool
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
	runCmd.Flags().StringVar(&ipfsrevproxy, "ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
	runCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Run through steps without pushing data to Powergate API")
	runCmd.Flags().StringVarP(&resultsOut, "results", "r", "results.json", "The location to store intermediate and final results.")
	runCmd.Flags().StringVar(&taskFolder, "folder", "", "Folder with organized tasks of directories or files")
	runCmd.Flags().Int64Var(&concurrent, "concurrent", 4, "Max concurrent tasks being processed")
	runCmd.Flags().Int64Var(&maxStagedBytes, "max-staged-bytes", 30000, "Maximum bytes of all tasks queued on staging")
	runCmd.Flags().Int64Var(&maxDealBytes, "max-deal-bytes", 24000, "Maximum bytes of a single deal")
	runCmd.Flags().Int64Var(&minDealBytes, "min-deal-bytes", 8000, "Minimum bytes of a single deal")
	runCmd.Flags().BoolVarP(&hiddenFiles, "all", "a", false, "Include hidden files & folders from top level folder")
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Run in non-interactive")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run in verbose mode")
	runCmd.Flags().BoolVarP(&retryErrors, "retry-errors", "f", false, "Retry tasks with error status")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a storage task pipeline",
	Long:  `Run a storage task pipeline to migrate large collections of data to Filecoin`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if maxStagedBytes < maxDealBytes {
			Fatal(fmt.Errorf("Max deal size (%d) is larger than max staging size (%d)", maxDealBytes, maxStagedBytes))
		}

		unfiltered, err := pathToTasks(taskFolder)
		checkErr(err)

		tasks := filterTasks(unfiltered, verbose)

		if !quiet {
			Message("New tasks %d", len(tasks))
		}

		// Ensure there aren't existing tasks in the same queue
		knownTasks, err := openResults(resultsOut)
		checkErr(err)
		if !quiet {
			Message("Existing tasks: %d", len(knownTasks))
		}
		tasks = mergeNewTasks(knownTasks, tasks)

		tasks = removeComplete(tasks)
		if !quiet {
			Message("Total pending tasks: %d", len(tasks))
		}

		if len(tasks) == 0 {
			return
		}

		// Init or update our intermediate results
		err = storeResults(resultsOut, tasks)
		checkErr(err)

		// Determine the max concurrent we need vs allowed
		if int64(len(tasks)) < concurrent {
			concurrent = int64(len(tasks))
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
		rc.storageConfig = storageConfig

		var progressWaitGroup sync.WaitGroup
		progress := uiprogress.New()
		progress.Start()
		progressBars := make(map[string](chan Task))

		taskUpdates := Run(tasks, rc)

		jobs := 0
		deals := 0
		errs := 0
		jobCt := 1
		for task := range taskUpdates {
			if _, found := progressBars[task.Name]; !found {
				progressBars[task.Name] = make(chan Task)
				progressWaitGroup.Add(1)
				go progressBar(progress, progressBars[task.Name], jobCt, len(tasks), &progressWaitGroup)
				jobCt++
			}
			if task.err != nil {
				errs++
			} else {
				if task.Stage > DryRunComplete {
					progressBars[task.Name] <- task
				}
				if task.Stage == Complete {
					jobs++
				}
			}
			tasks = updateTasks(tasks, task)
			err = storeResults(resultsOut, tasks)
			checkErr(err)
		}

		go func() {
			progressWaitGroup.Wait()
			for _, c := range progressBars {
				close(c)
			}
		}()

		progress.Stop()
		Message("%d %d %d", jobs, deals, errs)
	},
}

func getStorageConfig(config PipelineConfig) (*userPb.StorageConfig, error) {
	pow, err := client.NewClient(config.serverAddress)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, config.token)
	task, err := pow.StorageConfig.Default(ctx)
	if err != nil {
		return nil, err
	}
	return task.GetDefaultStorageConfig(), nil
}

func updateTasks(tasks []Task, update Task) []Task {
	for i, orig := range tasks {
		if orig.Path == update.Path {
			tasks[i] = update
			return tasks
		}
	}
	return tasks
}
