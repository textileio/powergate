package cmd

import (
	"fmt"
	"os"
	"strings"

	"encoding/csv"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	runCmd.Flags().Bool("dry-run", false, "Run through steps without pushing data to Powergate API")
	runCmd.Flags().String("ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
	runCmd.Flags().String("folder", "", "Folder with organized tasks of directories or files")
	runCmd.Flags().Int64("concurrent", 2, "Max concurrent tasks being processed")
	runCmd.Flags().Int64("maxStagedBytes", 4000000000, "Maximum bytes of all tasks queued on staging")
	runCmd.Flags().Int64("maxDealBytes", 300000000, "Maximum bytes of a single deal")
	runCmd.Flags().Int64("minDealBytes", 25000, "Minimum bytes of a single deal")
	runCmd.Flags().Bool("hidden", false, "Include hidden files & folders from top level folder")
	runCmd.Flags().String("jobs", "jobs.csv", "Output file for jobs results")
	runCmd.Flags().String("deals", "deals.csv", "Output file for deals results")
	runCmd.Flags().String("errors", "errors.csv", "Output file for errors results")
	runCmd.Flags().Bool("pipe", false, "Pipe all results to stdout")
	runCmd.Flags().Bool("debug", false, "Debug")

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
		maxStagedBytes := viper.GetInt64("maxStagedBytes")
		maxDealBytes := viper.GetInt64("maxDealBytes")
		if maxStagedBytes < maxDealBytes {
			Fatal(fmt.Errorf("Max deal size (%d) is larger than max staging size (%d)", maxDealBytes, maxStagedBytes))
		}

		// TODO: Implement "file" task input with a list of locations/urls
		folder := viper.GetString("folder")

		// TODO: Add file size filter
		alltsks, err := PathToTasks(folder)
		checkErr(err)

		Message("Found %d tasks", len(alltsks))

		tasks := alltsks[:0]

		hidden := viper.GetBool("hidden")
		minDealBytes := viper.GetInt64("minDealBytes")

		for _, tsk := range alltsks {
			if !hidden && strings.HasPrefix(tsk.name, ".") {
				Message("removing hidden %s", tsk.name)
				continue
			}
			if tsk.bytes < minDealBytes {
				Message("data too small %s (%d)", tsk.name, tsk.bytes)
				continue
			}
			if maxDealBytes < tsk.bytes {
				Message("data too large %s (%d)", tsk.name, tsk.bytes)
				continue
			}
			tasks = append(tasks, tsk)
		}

		Message("Removed %d tasks", len(alltsks)-len(tasks))

		concurrent := viper.GetInt64("concurrent")
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
			dryRun:          viper.GetBool("dry-run"),
		}

		res := Run(tasks, rc)

		if viper.GetBool("pipe") {
			jobs, deals, errs := resultsToStdOut(res)
			Success("Complete: %d tasks. %d errors %d deals\n", jobs+errs, errs, deals)
		} else {
			jobsFile := viper.GetString("jobs")
			dealsFile := viper.GetString("deals")
			errorsFile := viper.GetString("errors")
			jobs, deals, errs := resultsToCSV(res, jobsFile, dealsFile, errorsFile)
			Success("Complete: %d tasks. %d errors %d deals\n", jobs+errs, errs, deals)
		}

	},
}

func resultsToStdOut(rc chan TaskResult) (int, int, int) {
	jobs := 0
	deals := 0
	errs := 0
	for {
		select {
		case res, ok := <-rc:
			if ok {
				if res.err != nil {
					Message("Error: %s %s %s", res.task.path, res.stage, res.err.Error())
					errs++
				} else {
					Message("Job: %s %s %s", res.task.path, res.cid, res.jobId)

					for _, record := range res.records {
						Message("Deal: %s %s %s", res.task.path, res.cid, record.DealInfo.ProposalCid)
						deals++
					}
					jobs++
				}
			} else {
				return jobs, deals, errs
			}
		}
	}
}
func resultsToCSV(rc chan TaskResult, jobsFileName string, dealsFileName string, errFileName string) (int, int, int) {
	jobs := 0
	deals := 0
	errs := 0

	errFile, err := os.Create(errFileName)
	checkErr(err)
	errCsv := csv.NewWriter(errFile)

	resFile, err := os.Create(jobsFileName)
	checkErr(err)

	resCsv := csv.NewWriter(resFile)

	dealFile, err := os.Create(dealsFileName)
	checkErr(err)

	dealCsv := csv.NewWriter(dealFile)

	errCsv.Write([]string{
		"name",
		"path",
		"bytes",
		"cid",
		"jobId",
		"stage",
		"error",
	})
	resCsv.Write([]string{
		"name",
		"path",
		"bytes",
		"cid",
		"jobId",
		"stage",
	})

	dealCsv.Write([]string{
		"name",
		"path",
		"cid",
		"jobId",
		"address",
		"rootCid",
		"pending",
		"time",
		"miner",
		"proposalCid",
		"stateName",
	})

	for {
		select {
		case res, ok := <-rc:
			if ok {
				if res.err != nil {
					errCsv.Write([]string{
						res.task.name,
						res.task.path,
						fmt.Sprint(res.task.bytes),
						res.cid,
						res.jobId,
						res.stage,
						res.err.Error(),
					})
					Message("Error: %s", res.err.Error())
					errs++
				} else {
					resCsv.Write([]string{
						res.task.name,
						res.task.path,
						fmt.Sprint(res.task.bytes),
						res.cid,
						res.jobId,
						res.stage,
					})
					Message("Complete: %s %s", res.task.path, res.cid)

					for _, record := range res.records {
						dealCsv.Write([]string{
							res.task.name,
							res.task.path,
							res.cid,
							res.jobId,
							record.Address,
							record.RootCid,
							fmt.Sprint(record.Pending),
							fmt.Sprint(record.Time),
							record.DealInfo.Miner,
							record.DealInfo.ProposalCid,
							record.DealInfo.StateName,
						})
						deals++
					}
					jobs++
				}
			} else {
				errCsv.Flush()
				if err = errCsv.Error(); err != nil {
					NonFatal(err)
				}
				resCsv.Flush()
				if err = resCsv.Error(); err != nil {
					NonFatal(err)
				}
				dealCsv.Flush()
				if err = dealCsv.Error(); err != nil {
					NonFatal(err)
				}
				if err = errFile.Close(); err != nil {
					NonFatal(err)
				}
				if err = resFile.Close(); err != nil {
					NonFatal(err)
				}
				if err = dealFile.Close(); err != nil {
					NonFatal(err)
				}
				return jobs, deals, errs
			}
		}
	}
}

func PathToTasks(target string) ([]Task, error) {
	target = filepath.Clean(target)
	files, err := ioutil.ReadDir(target)
	if err != nil {
		return nil, err
	}

	tasks := []Task{}

	for _, f := range files {
		fullPath := filepath.Join(target, f.Name())
		if f.IsDir() {
			size, err := getDirSize(fullPath)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, Task{
				name:     f.Name(),
				path:     fullPath,
				bytes:    size,
				isDir:    true,
				isHidden: strings.HasPrefix(f.Name(), "."),
			})
			continue
		}
		tasks = append(tasks, Task{
			name:     f.Name(),
			path:     fullPath,
			bytes:    f.Size(),
			isDir:    false,
			isHidden: strings.HasPrefix(f.Name(), "."),
		})
	}

	return tasks, nil
}
