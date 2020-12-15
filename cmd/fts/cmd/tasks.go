package cmd

import (
	"sort"
	"sync"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// PipelineConfig config for task pipeline.
type PipelineConfig struct {
	token           string
	ipfsrevproxy    string
	serverAddress   string
	concurrent      int64
	minDealBytes    int64
	maxStagingBytes int64
	dryRun          bool
	mainnet         bool
	storageConfig   *userPb.StorageConfig
}

type fanoutConf struct {
	updates chan<- Task
	staging *StagingStatus
	pipe    PipelineConfig
	pipeWg  *sync.WaitGroup
}

// Start the task pipeline.
func Start(tasks []Task, pipe PipelineConfig) chan Task {
	pipeWg := sync.WaitGroup{}
	sc := StagingStatus{
		maxStagingBytes: pipe.maxStagingBytes,
		minDealBytes:    pipe.minDealBytes,
	}
	// updates chan for return to gui, pipe, and output.
	updates := make(chan Task)

	conf := fanoutConf{
		updates: updates,
		staging: &sc,
		pipe:    pipe,
		pipeWg:  &pipeWg,
	}
	pipeWg.Add(1)
	go fanout(tasks, conf)

	go func() {
		pipeWg.Wait()
		close(updates)
	}()

	return updates
}

func fanout(tasks []Task, conf fanoutConf) {
	defer conf.pipeWg.Done()
	readyToWatch := make(chan Task)
	complete := make(chan struct{})

	queueConf := queueConfig{
		watch:    readyToWatch,
		complete: complete,
		updates:  conf.updates,
		staging:  conf.staging,
		pipe:     conf.pipe,
		pipeWg:   conf.pipeWg,
	}

	conf.pipeWg.Add(1)
	go watcher(queueConf)

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
