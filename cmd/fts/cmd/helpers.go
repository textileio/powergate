package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
)

// Message prints a message to stdout.
func Message(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightBlack("> "+format), args...))
}

// Success prints a success message to stdout.
func Success(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.Cyan("> Success! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

// Success prints a success message to stdout.
func Warning(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightYellow("> Warning! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

// Fatal prints a fatal error to stdout, and exits immediately with
// error code 1.
func Fatal(err error, args ...interface{}) {
	NonFatal(err, args)
	os.Exit(1)
}

// Fatal prints a fatal error to stdout, and exits immediately with
// error code 1.
func NonFatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Println(aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
}

// RenderTable renders a table with header columns and data rows to writer.
func RenderTable(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorder(false)
	headersColors := make([]tablewriter.Colors, len(header))
	for i := range headersColors {
		headersColors[i] = tablewriter.Colors{tablewriter.FgHiBlackColor}
	}
	table.SetHeaderColor(headersColors...)
	table.AppendBulk(data)
	table.Render()
}

func checkErr(e error) {
	if e != nil {
		Fatal(e)
	}
}

func mustAuthCtx(ctx context.Context) context.Context {
	token := viper.GetString("token")
	if token == "" {
		Fatal(errors.New("must provide -t token"))
	}
	return context.WithValue(ctx, client.AuthKey, token)
}

func adminAuthCtx(ctx context.Context) context.Context {
	token := viper.GetString("admin-token")
	return context.WithValue(ctx, client.AdminKey, token)
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func mergeNewTasks(primary []Task, secondary []Task) []Task {
	for _, task := range secondary {
		primary = appendUniquePaths(primary, task)
	}
	return primary
}

func appendUniquePaths(tasks []Task, task Task) []Task {
	for _, orig := range tasks {
		if orig.Path == task.Path {
			return tasks
		}
	}
	return append(tasks, task)
}

func storeResults(target string, tasks []Task) error {
	target, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return err
	}
	file, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(target, file, 0644)
}

func openResults(target string) ([]Task, error) {
	target, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(target); err == nil {
		plan, _ := ioutil.ReadFile(target)
		var data []Task
		err = json.Unmarshal(plan, &data)
		return data, err
	} else if os.IsNotExist(err) {
		return make([]Task, 0), nil
	} else {
		return make([]Task, 0), err
	}
}

func resultsToStdOut(rc chan Task) (int, int, int) {
	jobs := 0
	deals := 0
	errs := 0
	for {
		select {
		case res, ok := <-rc:
			if ok {
				if res.err != nil {
					Message("Error: %s %s %s", res.Path, res.Stage, res.err.Error())
					errs++
				} else {
					Message("Job: %s %s %s", res.Path, res.CID, res.JobID)

					for _, record := range res.records {
						Message("Deal: %s %s %s", res.Path, res.CID, record.DealInfo.ProposalCid)
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

func resultsToCSV(rc chan Task, jobsFileName string, dealsFileName string, errFileName string) (int, int, int) {
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
						res.Name,
						res.Path,
						fmt.Sprint(res.Bytes),
						res.CID,
						res.JobID,
						string(res.Stage),
						res.err.Error(),
					})
					Message("Error: %s", res.err.Error())
					errs++
				} else {
					resCsv.Write([]string{
						res.Name,
						res.Path,
						fmt.Sprint(res.Bytes),
						res.CID,
						res.JobID,
						string(res.Stage),
					})
					Message("Complete: %s %s", res.Path, res.CID)

					for _, record := range res.records {
						dealCsv.Write([]string{
							res.Name,
							res.Path,
							res.CID,
							res.JobID,
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

func filterTasks(tsks []Task, verbose bool) []Task {
	tasks := tsks[:0]
	for _, tsk := range tsks {
		if !hiddenFiles && strings.HasPrefix(tsk.Name, ".") {
			if verbose {
				Message("removing hidden %s", tsk.Name)
			}
			continue
		}
		if tsk.Bytes < minDealBytes {
			if verbose {
				Message("data too small %s (%d)", tsk.Name, tsk.Bytes)
			}
			continue
		}
		if maxDealBytes < tsk.Bytes {
			if verbose {
				Message("data too large %s (%d)", tsk.Name, tsk.Bytes)
			}
			continue
		}
		tasks = append(tasks, tsk)
	}
	return tasks
}

func pathToTasks(target string) ([]Task, error) {
	target, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return nil, err
	}
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
				Name:     f.Name(),
				Path:     fullPath,
				Bytes:    size,
				IsDir:    true,
				IsHidden: strings.HasPrefix(f.Name(), "."),
				Stage:    stageToString[Init],
			})
			continue
		}
		tasks = append(tasks, Task{
			Name:     f.Name(),
			Path:     fullPath,
			Bytes:    f.Size(),
			IsDir:    false,
			IsHidden: strings.HasPrefix(f.Name(), "."),
			Stage:    stageToString[Init],
		})
	}

	return tasks, nil
}
