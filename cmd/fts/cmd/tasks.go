package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	client "github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// TaskStage is an int representing the task stage
type TaskStage int

const (
	DryRunComplete TaskStage = iota
	Init
	Staging
	StagingComplete
	PushingJob
	PushComplete
	DealsStarting
	DealsQueued
	DealsExecuting
	DealsComplete
	Complete
)

var stageToString = map[TaskStage]string{
	DryRunComplete:  "DryRunComplete",
	Init:            "Init",
	Staging:         "Staging",
	StagingComplete: "StagingComplete",
	PushingJob:      "PushingJob",
	PushComplete:    "PushComplete",
	DealsStarting:   "DealsStarting",
	DealsQueued:     "DealsQueued",
	DealsExecuting:  "DealsExecuting",
	DealsComplete:   "DealsComplete",
	Complete:        "Complete",
}

var stringToStage = map[string]TaskStage{
	"DryRunComplete":  DryRunComplete,
	"Init":            Init,
	"Staging":         Staging,
	"StagingComplete": StagingComplete,
	"PushingJob":      PushingJob,
	"PushComplete":    PushComplete,
	"DealsStarting":   DealsStarting,
	"DealsQueued":     DealsQueued,
	"DealsExecuting":  DealsExecuting,
	"DealsComplete":   DealsComplete,
	"Complete":        Complete,
}

func (t TaskStage) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(stageToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *TaskStage) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*t = stringToStage[j]
	return nil
}

type PipelineConfig struct {
	token           string
	ipfsrevproxy    string
	serverAddress   string
	concurrent      int64
	minDealBytes    int64
	maxStagingBytes int64
	dryRun          bool
	storageConfig   *userPb.StorageConfig
}

func dryRunStep() {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(250)))
}

type fanoutConf struct {
	updates chan<- Task
	wg      *sync.WaitGroup
	staging *StagingLimit
	pipe    PipelineConfig
}

func fanout(tasks []Task, conf fanoutConf) {
	defer conf.wg.Done()
	readyToWatch := make(chan Task)
	complete := make(chan struct{})

	queueConf := queueConfig{
		watch:    readyToWatch,
		complete: complete,
		updates:  conf.updates,
		staging:  conf.staging,
		pipe:     conf.pipe,
		wg:       conf.wg,
	}
	conf.wg.Add(1)
	go watchJobs(queueConf)

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Stage > tasks[j].Stage
	})
	last := false
	for i, task := range tasks {
		if i == len(tasks)-1 {
			last = true
		}
		queue(task, queueConf, last)
	}
}

// Run this thing
func Run(tasks []Task, pipe PipelineConfig) chan Task {
	wg := sync.WaitGroup{}
	sc := StagingLimit{
		maxStagingBytes: pipe.maxStagingBytes,
		minDealBytes:    pipe.minDealBytes,
	}
	updates := make(chan Task)

	conf := fanoutConf{
		updates: updates,
		wg:      &wg,
		staging: &sc,
		pipe:    pipe,
	}
	wg.Add(1)
	go fanout(tasks, conf)

	go func() {
		wg.Wait()
		close(updates)
	}()

	return updates
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
