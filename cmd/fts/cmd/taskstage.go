package cmd

import (
	"bytes"
	"encoding/json"
)

// TaskStage is an int representing the task stage.
type TaskStage int

const (
	// DryRunComplete is the end status for dry-run tasks.
	DryRunComplete TaskStage = iota
	// Init is a newly created task.
	Init
	// Staging is a task being uploaded to staging.
	Staging
	// StagingComplete is a task that exists on staging.
	StagingComplete
	// PushingJob is a task config being pushed to Pow.
	PushingJob
	// PushComplete is a task in the job queue of the Pow.
	PushComplete
	// DealsStarting is when a job when the watcher is waiting for an update.
	DealsStarting
	// DealsQueued when the job is in the Powergate queue.
	DealsQueued
	// DealsExecuting is when miners have been found and the deal is processing.
	DealsExecuting
	// DealsComplete means all deals have finished.
	DealsComplete
	// Complete task.
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

// MarshalJSON turns our TaskStage to valid json.
func (t TaskStage) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(stageToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON turns json to TaskStage.
func (t *TaskStage) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*t = stringToStage[j]
	return nil
}
