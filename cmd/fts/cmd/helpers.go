package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
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

// Warning prints a warning message to stdout.
func Warning(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightYellow("> Warning! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

// Fatal prints a fatal error to stdout, and exits immediately with
// error code 1.
func Fatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Println(aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
	os.Exit(1)
}

// NonFatal prints a fatal error to stdout.
func NonFatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Println(aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
}

func checkErr(e error) {
	if e != nil {
		Fatal(e)
	}
}

// Get the on-disk size of a directory.
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

// Remote tasks with Complete status.
func removeComplete(tasks []Task) []Task {
	result := []Task{}
	for _, task := range tasks {
		if task.Stage != Complete {
			result = append(result, task)
		}
	}
	return result
}

// Remote tasks with an error.
func cleanErrors(tasks []Task, retry bool) []Task {
	result := []Task{}
	for _, task := range tasks {
		if task.Error == "" {
			result = append(result, task)
		} else if retry {
			task.Stage = Init
			task.Error = ""
			result = append(result, task)
		}
	}
	return result
}

// Merge two task lists without duplicates.
func mergeNewTasks(primary []Task, secondary []Task) []Task {
	for _, task := range secondary {
		primary = appendUniquePaths(primary, task)
	}
	return primary
}

// Ensures no duplicate tasks based on path.
func appendUniquePaths(tasks []Task, task Task) []Task {
	for _, orig := range tasks {
		if orig.Path == task.Path {
			return tasks
		}
	}
	return append(tasks, task)
}

// Stores the results json as a file.
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

// Opens the json results output.
func openResults(target string) ([]Task, error) {
	target, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(target); err == nil {
		plan, _ := ioutil.ReadFile(target)
		var tasks []Task
		err = json.Unmarshal(plan, &tasks)
		return tasks, err
	} else if os.IsNotExist(err) {
		return make([]Task, 0), nil
	} else {
		return make([]Task, 0), err
	}
}

// Filter hidden files/folders out of task slice.
func filterTasks(tasks []Task) []Task {
	result := tasks[:0]
	for _, tsk := range tasks {
		if !hiddenFiles && strings.HasPrefix(tsk.Name, ".") {
			continue
		}
		if tsk.Bytes < minDealBytes {
			continue
		}
		if maxDealBytes < tsk.Bytes {
			continue
		}
		result = append(result, tsk)
	}
	return result
}

// Converts a directory into a slice of Tasks.
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
				Stage:    Init,
			})
			continue
		}
		tasks = append(tasks, Task{
			Name:     f.Name(),
			Path:     fullPath,
			Bytes:    f.Size(),
			IsDir:    false,
			IsHidden: strings.HasPrefix(f.Name(), "."),
			Stage:    Init,
		})
	}

	return tasks, nil
}

// gets the default storage config from powergate.
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

// overwrites a Task in a slice.
func updateTasks(tasks []Task, update Task) ([]Task, bool) {
	for i, orig := range tasks {
		if orig.Path == update.Path {
			change := orig.Stage != update.Stage || update.Error != ""
			tasks[i] = update
			return tasks, change
		}
	}
	return tasks, false
}

func stageData(task Task, pow *client.Client, config PipelineConfig) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, config.token)
	if task.IsDir {
		return pow.Data.StageFolder(ctx, config.ipfsrevproxy, task.Path)
	}
	f, err := os.Open(task.Path)
	if err != nil {
		return "", err
	}
	stageRes, err := pow.Data.Stage(ctx, f)
	if err != nil {
		return "", err
	}
	err = f.Close()
	return stageRes.Cid, err
}

// applyConfig applys either default or given config to a CID.
func applyConfig(task Task, pow *client.Client, config PipelineConfig) (*userPb.ApplyStorageConfigResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, config.token)
	options := []client.ApplyOption{}
	if task.storageConfig != nil {
		options = append(options, client.WithStorageConfig(task.storageConfig))
	} else if config.storageConfig != nil {
		options = append(options, client.WithStorageConfig(config.storageConfig))
	}
	return pow.StorageConfig.Apply(ctx, task.CID, options...)
}

func dryRunStep() {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(250)))
}
