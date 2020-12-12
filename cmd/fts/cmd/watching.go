package cmd

import (
	"context"
	"sync"
	"time"

	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

type watchJobsConfig struct {
	watch    chan Task
	updates  chan<- Task
	staging  *StagingLimit
	pipe     PipelineConfig
	complete chan struct{}
	wg       *sync.WaitGroup
}

// timed loop that will check for deal updates
func watchJobs(conf queueConfig) {
	defer conf.wg.Done()
	watching := []Task{}
	delay := time.Second * 3

	receivedAllTasks := false

	for range time.Tick(delay) {
		incomplete := []Task{}
		summary, err := getStorageSummary(conf)
		queued := summary.QueuedStorageJobs
		executing := summary.ExecutingStorageJobs
		final := summary.LatestFinalStorageJobs
		if err != nil {
			continue
		}
		for _, task := range watching {
			if task.Stage < DealsComplete {
				newStage := DealsStarting
				task.Stage = newStage
				conf.updates <- task
				if conf.pipe.dryRun {
					newStage = Complete
					dryRunStep()
				} else {
					deals := make([]*userPb.StorageJob, 0)
					for _, job := range queued {
						if job.Cid == task.CID {
							deals = append(deals, job)
							newStage = DealsQueued
						}
					}
					for _, job := range executing {
						if job.Cid == task.CID {
							deals = append(deals, job)
							newStage = DealsExecuting
						}
					}
					completed := 0
					for _, job := range final {
						if job.Cid == task.CID {
							deals = append(deals, job)
							newStage = DealsComplete
							completed++
						}
					}
					task.Deals = deals
					if completed == len(deals) {
						conf.staging.Done(task.Bytes)
						newStage = Complete
					} else {
						incomplete = append(incomplete, task)
					}
				}
				task.Stage = newStage
				conf.updates <- task
			}
			if conf.pipe.dryRun {
				conf.staging.Done(task.Bytes)
				task.Stage = DryRunComplete
				conf.updates <- task
			}
		}
		watching = append(incomplete, updateWatching(conf.watch)...)
		if !receivedAllTasks {
			receivedAllTasks = checkReceivedAll(conf.complete)
		}
		if receivedAllTasks && len(watching) == 0 {
			return
		}
	}
}

// grabs any new tasks that are ready to watch
func updateWatching(tasks chan Task) []Task {
	watching := []Task{}
	for {
		select {
		case task := <-tasks:
			watching = append(watching, task)
		default:
			return watching
		}
	}
}

// complete is fired after the final step has been staged.
func checkReceivedAll(complete chan struct{}) bool {
	for {
		select {
		case <-complete:
			return true
		default:
			return false
		}
	}
}

func getStorageSummary(conf queueConfig) (*userPb.StorageJobsSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*2)
	defer cancel()
	ctx = context.WithValue(ctx, client.AuthKey, conf.pipe.token)

	pow, err := client.NewClient(conf.pipe.serverAddress)
	if err != nil {
		return nil, err
	}
	summary, err := pow.StorageJobs.Summary(ctx)
	if err != nil {
		return nil, err
	}
	return summary, nil
}
