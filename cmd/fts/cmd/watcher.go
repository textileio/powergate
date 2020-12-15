package cmd

import (
	"context"
	"time"

	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// timed loop that will check for deal updates.
func watcher(conf queueConfig) {
	defer conf.pipeWg.Done()
	watching := []Task{}
	wait := 2 * time.Second
	maxWait := 2 * time.Minute
	if conf.pipe.mainnet {
		maxWait = 5 * time.Minute
	}
	receivedAllTasks := false

	for range time.Tick(wait) {
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
		// Allows any empty jobs or dry-runs to clean up fast.
		if wait < maxWait {
			wait = wait * 2
		}
		if receivedAllTasks && len(watching) == 0 {
			return
		}
	}
}

// grabs any new tasks that are ready to watch.
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
