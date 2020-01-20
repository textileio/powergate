package ask

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	keyAskStatus, _ = tag.NewKey("askstatus")

	mFullRefreshDuration = stats.Int64("askindex/fullrefresh-duration", "Duration of Full StorageAsk refresh", "s")
	mFullRefreshProgress = stats.Float64("askindex/fullrefresh-progress", "Full refresh progress", "By")
	mAskQueryResult      = stats.Int64("askindex/queryask-result", "Ask query results", "By")

	vFullRefreshTime = &view.View{
		Name:        "askindex/fullrefresh-duration",
		Measure:     mFullRefreshDuration,
		Description: "Full StorageAsk refresh duration",
		Aggregation: view.LastValue(),
	}
	vFullRefreshProgress = &view.View{
		Name:        "askindex/fullrefresh-progress",
		Measure:     mFullRefreshProgress,
		Description: "Full refresh progress",
		Aggregation: view.LastValue(),
	}
	vAskQueryResult = &view.View{
		Name:        "askindex/queryask-result",
		Measure:     mAskQueryResult,
		Description: "Query-Ask results",
		TagKeys:     []tag.Key{keyAskStatus},
		Aggregation: view.LastValue(),
	}
	views = []*view.View{vFullRefreshTime, vAskQueryResult, vFullRefreshProgress}
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
