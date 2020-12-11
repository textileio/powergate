package cmd

import (
	"errors"
	"fmt"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
)

func init() {
	runCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Run through steps without pushing data to Powergate API")
	runCmd.Flags().StringVarP(&resultsOut, "results", "r", "results.json", "The location to store intermediate and final results.")
	runCmd.Flags().String("ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
	runCmd.Flags().StringVar(&taskFolder, "folder", "", "Folder with organized tasks of directories or files")
	runCmd.Flags().Int64Var(&concurrent, "concurrent", 4, "Max concurrent tasks being processed")
	runCmd.Flags().Int64Var(&maxStagedBytes, "max-staged-bytes", 30000, "Maximum bytes of all tasks queued on staging")
	runCmd.Flags().Int64Var(&maxDealBytes, "max-deal-bytes", 10000, "Maximum bytes of a single deal")
	runCmd.Flags().Int64Var(&minDealBytes, "min-deal-bytes", 8699, "Minimum bytes of a single deal")
	runCmd.Flags().BoolVarP(&hiddenFiles, "all", "a", false, "Include hidden files & folders from top level folder")
	runCmd.Flags().String("jobs", "jobs.csv", "Output file for jobs results")
	runCmd.Flags().String("deals", "deals.csv", "Output file for deals results")
	runCmd.Flags().String("errors", "errors.csv", "Output file for errors results")
	runCmd.Flags().Bool("pipe", false, "Pipe all results to stdout")
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Run in non-interactive")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Run in verbose mode")

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
			Message("Found %d tasks (filtered %d)", len(tasks), len(unfiltered)-len(tasks))
		}

		// Prompt to continue
		if len(tasks) > 0 && !quiet {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Continue with %d tasks", len(tasks)),
				IsConfirm: true,
				Default:   "y",
			}
			result, _ := prompt.Run()
			if err != nil && err != errors.New("") {
				checkErr(err)
			}
			if result == "N" || result == "n" {
				return
			}
		}

		// Ensure there aren't existing tasks in the same queue
		knownTasks, err := openResults(resultsOut)
		checkErr(err)
		if len(knownTasks) > 0 && !quiet {
			Message("Previous job exists (%d tasks)", len(knownTasks))
			prompt := promptui.Select{
				Label: "Continue or discard old job tasks",
				Items: []string{
					"Continue",
					"Discard",
				},
			}

			choice, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}
			if choice != 0 {
				knownTasks = make([]Task, 0)
			}
		}
		tasks = mergeNewTasks(knownTasks, tasks)

		// Init or update our intermediate results
		err = storeResults(resultsOut, tasks)
		checkErr(err)

		// Determine the max concurrent we need vs allowed
		if int64(len(tasks)) < concurrent {
			concurrent = int64(len(tasks))
		}

		rc := RunConfig{
			token:           viper.GetString("token"),
			ipfsrevproxy:    viper.GetString("ipfsrevproxy"),
			serverAddress:   viper.GetString("serverAddress"),
			maxStagingBytes: maxStagedBytes,
			minDealBytes:    minDealBytes,
			concurrent:      concurrent,
			dryRun:          dryRun,
		}

		bar := uiprogress.AddBar(len(tasks)).AppendCompleted().PrependElapsed()
		bar.PrependFunc(func(b *uiprogress.Bar) string {
			return strutil.Resize(fmt.Sprintf("Task (%d/%d)", b.Current(), len(tasks)), 22)
		})
		uiprogress.Start()

		updates := Run(tasks, rc)
		jobs := 0
		deals := 0
		errs := 0
		for {
			select {
			case res, ok := <-updates:
				if ok {
					if res.err != nil {
						// Message("Error: %s %s %s", res.Path, res.Stage, res.err.Error())
						errs++
					} else {
						// Message("Job: %s %s %s", res.Path, res.CID, res.JobID)

						// for _, record := range res.records {
						// 	// Message("Deal: %s %s %s", res.Path, res.CID, record.DealInfo.ProposalCid)
						// 	deals++
						// }
						if res.Stage == "Complete" {
							// Message("%s: %s", res.Stage, res.Name)
							bar.Incr()
							jobs++
						}
					}
				} else {
					uiprogress.Stop()
					Message("%d %d %d", jobs, deals, errs)
					return
				}
			}
		}
	},
}
