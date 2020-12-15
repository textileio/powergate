package cmd

import (
	"bytes"
	"encoding/json"
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
