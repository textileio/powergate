package cmd

import (
	"fmt"
	"io"
	"sync"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/olekukonko/tablewriter"
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
		err := bar.Set(int(task.Stage))
		checkErr(err)
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

// renderTable renders a table with header columns and data rows to writer.
func renderTable(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorder(false)
	headersColors := make([]tablewriter.Colors, len(header))
	for i := range headersColors {
		headersColors[i] = tablewriter.Colors{tablewriter.FgHiBlackColor}
	}
	table.SetHeaderColor(headersColors...)
	table.AppendBulk(data)
	table.Render()
}
