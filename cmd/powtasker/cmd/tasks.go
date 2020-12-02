package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	client "github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

type Task struct {
	name     string
	path     string
	bytes    int64
	isDir    bool
	isHidden bool
}

type TaskResult struct {
	task    Task
	cid     string
	jobId   string
	stage   string
	state   *client.WatchStorageJobsEvent
	records []*userPb.StorageDealRecord
	err     error
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

func Run(tasks []Task, config RunConfig) chan TaskResult {

	sc := StagingConfig{
		maxStagingBytes: config.maxStagingBytes,
		minDealBytes:    config.minDealBytes,
	}

	// wg is used to force the application wait for all goroutines to finish before exiting.
	wg := sync.WaitGroup{}
	// ch is a buffered channel for processing tasks
	ch := make(chan bool, config.concurrent)

	// res is a buffered channel for receiving results
	res := make(chan TaskResult, len(tasks))

	// How many jobs to expect
	wg.Add(len(tasks))

	for _, task := range tasks {
		go ProcessTask(task, config, ch, res, &wg, &sc)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}

func ProcessTask(task Task, config RunConfig, ch chan bool, res chan<- TaskResult, wg *sync.WaitGroup, sc *StagingConfig) {
	defer wg.Done()

	// Fill in the channel buffer slot.
	ch <- true

	taskResult := TaskResult{
		task:  task,
		stage: "waiting",
	}

	sendErr := func(err error) {
		taskResult.err = err
		res <- taskResult
	}

	for {
		// check available space
		ready, err := sc.Ready(task.bytes)
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

		Message("Starting storage of %s.", task.path)
		taskResult.stage = "started"
		// TODO: allow user to set target at input
		pow, err := client.NewClient(config.serverAddress)
		if err != nil {
			sc.Done(task.bytes)
			sendErr(err)
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()
		ctx = context.WithValue(ctx, client.AuthKey, config.token)

		if config.dryRun {
			taskResult.cid = "dry-run"
			taskResult.stage = "dry-run"
			sc.Done(task.bytes)
			res <- taskResult
			break
		}

		if task.isDir {
			taskResult.cid, err = pow.Data.StageFolder(ctx, config.ipfsrevproxy, task.path)
			if err != nil {
				sc.Done(task.bytes)
				sendErr(err)
				break
			}
		} else {
			f, err := os.Open(task.path)
			if err != nil {
				sc.Done(task.bytes)
				sendErr(err)
				break
			}
			stageRes, err := pow.Data.Stage(ctx, f)
			if err != nil {
				sc.Done(task.bytes)
				sendErr(err)
				break
			}
			taskResult.cid = stageRes.Cid
			if err := f.Close(); err != nil {
				sc.Done(task.bytes)
				sendErr(err)
				break
			}
		}

		taskResult.stage = "stagged"

		if config.dryRun {
			taskResult.stage = "dry-run"
			sc.Done(task.bytes)
			break
		}

		applyRes, err := pow.StorageConfig.Apply(ctx, taskResult.cid)
		if err != nil {
			sc.Done(task.bytes)
			sendErr(err)
			break
		}
		taskResult.jobId = applyRes.JobId

		taskResult.stage = "queued"

		Message("Waiting to complete deals for %s.", task.path)

		storage, err := watchStorage(pow, config.token, taskResult.jobId)
		if err != nil {
			sc.Done(task.bytes)
			sendErr(err)
			break
		}
		taskResult.state = storage

		Message("Completed storage of %s.", task.path)

		var opts []client.DealRecordsOption
		opts = append(opts, client.WithDataCids(taskResult.cid))
		opts = append(opts, client.WithIncludeFinal(true))
		opts = append(opts, client.WithIncludePending(false))
		complete, err := pow.Deals.StorageDealRecords(ctx, opts...)
		if err != nil {
			sc.Done(task.bytes)
			sendErr(err)
			break
		}
		taskResult.records = complete.Records
		sc.Done(task.bytes)
		res <- taskResult
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
