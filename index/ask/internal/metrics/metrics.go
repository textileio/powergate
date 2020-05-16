package metrics

import (
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	TagAskStatus, _ = tag.NewKey("askstatus")

	MFullRefreshDuration = stats.Int64("askindex/fullrefresh-duration", "Duration of Full StorageAsk refresh", "s")
	MFullRefreshProgress = stats.Float64("askindex/fullrefresh-progress", "Full refresh progress", "By")
	MAskQueryResult      = stats.Int64("askindex/queryask-result", "Ask query results", "By")

	vFullRefreshTime = &view.View{
		Name:        "askindex/fullrefresh-duration",
		Measure:     MFullRefreshDuration,
		Description: "Full StorageAsk refresh duration",
		Aggregation: view.LastValue(),
	}
	vFullRefreshProgress = &view.View{
		Name:        "askindex/fullrefresh-progress",
		Measure:     MFullRefreshProgress,
		Description: "Full refresh progress",
		Aggregation: view.LastValue(),
	}
	vAskQueryResult = &view.View{
		Name:        "askindex/queryask-result",
		Measure:     MAskQueryResult,
		Description: "Query-Ask results",
		TagKeys:     []tag.Key{TagAskStatus},
		Aggregation: view.LastValue(),
	}
	views = []*view.View{vFullRefreshTime, vAskQueryResult, vFullRefreshProgress}
)

func Init() error {
	if err := view.Register(views...); err != nil {
		return fmt.Errorf("Failed to register views: %v", err)
	}
	return nil
}
