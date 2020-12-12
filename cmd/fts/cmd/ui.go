package cmd

import (
	"fmt"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func getStandardBar(progress *uiprogress.Progress, width int) *uiprogress.Bar {
	bar := progress.AddBar(width)
	bar.PrependCompleted()
	bar.Width = 30
	return bar
}

func progressBar(progress *uiprogress.Progress, updates chan Task, job int, total int, pwg *sync.WaitGroup) {
	defer pwg.Done()
	var name string
	var stage string
	bar := getStandardBar(progress, int(Complete))
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(fmt.Sprintf("%s (%d/%d) %s", name, job, total, stage), 30)
	})
	for task := range updates {
		name, stage = staticNameStage(task)
		bar.Set(int(task.Stage))
	}
}

func staticNameStage(task Task) (string, string) {
	name := task.Name
	stage := stageToString[task.Stage]
	if task.Stage == DryRunComplete {
		stage = stageToString[Complete]
	}
	if task.Error != "" {
		stage = fmt.Sprintf("Error(%s)", stage)
	}
	return name, stage
}
