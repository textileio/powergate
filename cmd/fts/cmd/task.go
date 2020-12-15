package cmd

import (
	"sync"
	"time"

	client "github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// Task is an individual data collection to be stored on Filecoin.
type Task struct {
	Name          string               `json:"name,omitempty"`
	Path          string               `json:"path,omitempty"`
	Bytes         int64                `json:"bytes,omitempty"`
	IsDir         bool                 `json:"isDir,omitempty"`
	IsHidden      bool                 `json:"isHidden,omitempty"`
	CID           string               `json:"cid,omitempty"`
	JobID         string               `json:"jobID,omitempty"`
	Stage         TaskStage            `json:"stage,omitempty"`
	Error         string               `json:"error,omitempty"`
	Deals         []*userPb.StorageJob `json:"deals,omitempty"`
	err           error
	storageConfig *userPb.StorageConfig
}

type queueConfig struct {
	watch    chan Task
	updates  chan<- Task
	complete chan struct{}
	staging  *StagingStatus
	pipeWg   *sync.WaitGroup
	pipe     PipelineConfig
}

func queue(task Task, conf queueConfig, isLast bool) {
	waitTimeout := 1 * time.Second
	for {
		// check available space.
		ready, err := conf.staging.Ready(task.Bytes)
		if err != nil {
			task.Error = err.Error()
			task.err = err
			conf.updates <- task
			conf.pipeWg.Done()
			return
		}
		// wait for space.
		if !ready {
			time.Sleep(waitTimeout)
			waitTimeout = waitTimeout * 2
			continue
		}
		// start processing.
		conf.pipeWg.Add(1)
		go process(task, conf, isLast)
		return
	}
}

func process(task Task, conf queueConfig, isLast bool) {
	defer func() {
		if isLast {
			conf.complete <- struct{}{}
		}
		conf.pipeWg.Done()
	}()

	sendErr := func(err error) {
		task.Error = err.Error()
		task.err = err
		conf.staging.Done(task.Bytes)
		conf.updates <- task
	}
	sendStage := func(stage TaskStage) {
		task.Stage = stage
		conf.updates <- task
	}

	pow, err := client.NewClient(conf.pipe.serverAddress)
	if err != nil {
		sendErr(err)
		return
	}

	if task.Stage < StagingComplete {
		sendStage(Staging)
		if conf.pipe.dryRun {
			dryRunStep()
		} else {
			cid, err := stageData(task, pow, conf.pipe)
			task.CID = cid
			if err != nil {
				sendErr(err)
				return
			}
		}
		sendStage(StagingComplete)
	}

	if task.Stage < PushComplete {
		sendStage(PushingJob)
		if conf.pipe.dryRun {
			dryRunStep()
		} else {
			applyRes, err := applyConfig(task, pow, conf.pipe)
			if err != nil {
				sendErr(err)
				return
			}
			task.JobID = applyRes.JobId
		}
		sendStage(PushComplete)
	}

	conf.watch <- task
}
