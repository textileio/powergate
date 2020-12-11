package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	client "github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// TaskStage is an int representing the task stage
type TaskStage int

const (
	SkippedDryRun TaskStage = iota
	Init
	Queued
	Starting
	Staged
	ProcessingDeals
	Complete
)

var stageToString = map[TaskStage]string{
	SkippedDryRun:   "SkippedDryRun",
	Init:            "Init",
	Queued:          "Queued",
	Starting:        "Starting",
	Staged:          "Staged",
	ProcessingDeals: "ProcessingDeals",
	Complete:        "Complete",
}

// Task is an individual data collection to be stored on Filecoin
type Task struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	Bytes    int64  `json:"bytes,omitempty"`
	IsDir    bool   `json:"isDir,omitempty"`
	IsHidden bool   `json:"isHidden,omitempty"`
	CID      string `json:"cid,omitempty"`
	JobID    string `json:"jobID,omitempty"`
	Stage    string `json:"stage,omitempty"`
	Error    string `json:"error,omitempty"`
	state    *client.WatchStorageJobsEvent
	records  []*userPb.StorageDealRecord
	err      error
}

type RunConfig struct {
	token           string
	ipfsrevproxy    string
	serverAddress   string
	concurrent      int64
	minDealBytes    int64
	maxStagingBytes int64
	dryRun          bool
}

var steps = []string{
	"downloading source",
	"installing deps",
	"compiling",
	"packaging",
	"seeding database",
	"deploying",
	"staring servers",
}

func stage(task Task, pow *client.Client, config RunConfig) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
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

func process(task Task, updates chan<- Task, wg *sync.WaitGroup, staging *StagingConfig, config RunConfig) {
	defer wg.Done()

	sendErr := func(err error) {
		task.Error = err.Error()
		task.err = err
		updates <- task
	}
	sendStage := func(stage string) {
		task.Stage = stage
		updates <- task
	}

	bar := uiprogress.AddBar(len(steps)).AppendCompleted().PrependElapsed()
	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(fmt.Sprintf("%s  %s", task.Name, task.Stage), 22)
	})

	sendStage("Connecting")
	pow, err := client.NewClient(config.serverAddress)
	if err != nil {
		sendErr(err)
		return
	}

	sendStage("Staging")

	if config.dryRun {
		rand.Seed(500)
		for bar.Incr() {
			rn := rand.Intn(500)
			time.Sleep(time.Millisecond * time.Duration(rn))
			sendStage(fmt.Sprintf("%s-%d", "DryRun", rn))
		}
		staging.Done(task.Bytes)
		task.Stage = "Complete"
		updates <- task
		return
	}

	cid, err := stage(task, pow, config)
	task.CID = cid
	if err != nil {
		sendErr(err)
		return
	}

	rn := rand.Intn(5000)
	time.Sleep(time.Millisecond * time.Duration(rn))
	staging.Done(task.Bytes)
	task.Stage = "Complete"
	updates <- task
}

func queue(task Task, updates chan<- Task, wg *sync.WaitGroup, staging *StagingConfig, config RunConfig) {
	for {
		// check available space
		ready, err := staging.Ready(task.Bytes)
		if err != nil {
			task.Error = err.Error()
			task.err = err
			updates <- task
			wg.Done()
			return
		}
		// wait for space
		if !ready {
			timeout := 1000 * time.Millisecond
			time.Sleep(timeout)
			continue
		}
		// start processing
		go process(task, updates, wg, staging, config)
		return
	}
}

func fanout(tasks []Task, updates chan<- Task, wg *sync.WaitGroup, staging *StagingConfig, config RunConfig) {
	for _, task := range tasks {
		queue(task, updates, wg, staging, config)
	}
	wg.Done()
}

// Run this thing
func Run(tasks []Task, config RunConfig) chan Task {

	sc := StagingConfig{
		maxStagingBytes: config.maxStagingBytes,
		minDealBytes:    config.minDealBytes,
	}
	wg := sync.WaitGroup{}
	wg.Add(len(tasks) + 1)
	updates := make(chan Task)

	go fanout(tasks, updates, &wg, &sc, config)

	go func() {
		wg.Wait()
		close(updates)
	}()

	return updates
}

func RunW(tasks []Task, config RunConfig) chan Task {

	sc := StagingConfig{
		maxStagingBytes: config.maxStagingBytes,
		minDealBytes:    config.minDealBytes,
	}

	// wg is used to force the application wait for all goroutines to finish before exiting.
	wg := sync.WaitGroup{}
	// ch is a buffered channel for processing tasks
	ch := make(chan bool, config.concurrent)

	count := len(tasks)
	// res is a buffered channel for receiving results
	res := make(chan Task, count)

	// How many jobs to expect
	wg.Add(count)

	for _, task := range tasks {
		go ProcessTask(task, config, ch, res, &wg, &sc)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}

// ProcessTask runs the task
func ProcessTask(task Task, config RunConfig, ch chan bool, res chan<- Task, wg *sync.WaitGroup, sc *StagingConfig) {
	defer wg.Done()

	// Fill in the channel buffer slot.
	ch <- true

	task.Stage = stageToString[Queued]

	sendErr := func(err error) {
		task.Error = err.Error()
		task.err = err
		res <- task
	}

	sendRes := func(stage string) {
		task.Stage = stage
		res <- task
	}

	for {
		// check available space
		ready, err := sc.Ready(task.Bytes)
		if err != nil {
			sendErr(err)
			break
		}
		// wait for space
		if !ready {
			timeout := 1000 * time.Millisecond
			time.Sleep(timeout)
			continue
		}

		Message("Starting storage of %s.", task.Path)
		task.Stage = stageToString[Starting]
		// TODO: allow user to set target at input
		pow, err := client.NewClient(config.serverAddress)
		if err != nil {
			sc.Done(task.Bytes)
			sendErr(err)
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()
		ctx = context.WithValue(ctx, client.AuthKey, config.token)

		if config.dryRun {
			slp := time.Millisecond * time.Duration(rand.Intn(4000))
			time.Sleep(slp)
			Message("%d sleep", slp)
			task.CID = "dry-run"
			sc.Done(task.Bytes)
			sendRes("SkippedDryRun")
			break
		} else {
			Message("nope")
		}

		if task.IsDir {
			task.CID, err = pow.Data.StageFolder(ctx, config.ipfsrevproxy, task.Path)
			if err != nil {
				sc.Done(task.Bytes)
				sendErr(err)
				break
			}
		} else {
			f, err := os.Open(task.Path)
			if err != nil {
				sc.Done(task.Bytes)
				sendErr(err)
				break
			}
			stageRes, err := pow.Data.Stage(ctx, f)
			if err != nil {
				sc.Done(task.Bytes)
				sendErr(err)
				break
			}
			task.CID = stageRes.Cid
			if err := f.Close(); err != nil {
				sc.Done(task.Bytes)
				sendErr(err)
				break
			}
		}

		task.Stage = stageToString[Staged]

		applyRes, err := pow.StorageConfig.Apply(ctx, task.CID)
		if err != nil {
			sc.Done(task.Bytes)
			sendErr(err)
			break
		}
		task.JobID = applyRes.JobId

		task.Stage = stageToString[ProcessingDeals]

		Message("Waiting to complete deals for %s.", task.Path)

		storage, err := watchStorage(pow, config.token, task.JobID)
		if err != nil {
			sc.Done(task.Bytes)
			sendErr(err)
			break
		}
		task.state = storage

		Message("Completed storage of %s.", task.Path)

		var opts []client.DealRecordsOption
		opts = append(opts, client.WithDataCids(task.CID))
		opts = append(opts, client.WithIncludeFinal(true))
		opts = append(opts, client.WithIncludePending(false))
		complete, err := pow.Deals.StorageDealRecords(ctx, opts...)
		if err != nil {
			sc.Done(task.Bytes)
			sendErr(err)
			break
		}
		task.records = complete.Records
		sc.Done(task.Bytes)
		sendRes("Complete")
		break
	}
	// Free the channel buffer slot.
	<-ch
}

func watchStorage(pow *client.Client, token string, jobId string) (*client.WatchStorageJobsEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, token)

	wch := make(chan client.WatchStorageJobsEvent)

	err := pow.StorageJobs.Watch(ctx, wch, jobId)
	if err != nil {
		return nil, err
	}

	for {
		event, ok := <-wch
		if !ok {
			break
		}
		if event.Err != nil {
			return nil, event.Err
		}
		processing := false
		if event.Res.StorageJob.Status == userPb.JobStatus_JOB_STATUS_EXECUTING ||
			event.Res.StorageJob.Status == userPb.JobStatus_JOB_STATUS_QUEUED {
			processing = true
		}
		if processing && event.Err == nil {
			continue
		}
		return &event, nil
	}
	return nil, fmt.Errorf("Job watch returned incorrectly")
}
