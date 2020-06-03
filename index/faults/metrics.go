package faults

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	mRefreshProgress = stats.Float64("indexslashing/refresh-progress", "Refresh progress", "By")
	mRefreshDuration = stats.Int64("indexslashing/full-refresh-duration", "Duration of full-refresh", "s")
	mUpdatedHeight   = stats.Int64("indexslashing/updated-height", "Last updated height", "By")

	vRefreshProgress = &view.View{
		Name:        "indexslashing/refresh-progress",
		Measure:     mRefreshProgress,
		Description: "Refresh progress",
		Aggregation: view.LastValue(),
	}
	vRefreshDuration = &view.View{
		Name:        "indexslashing/refresh-duration",
		Measure:     mRefreshDuration,
		Description: "Refresh duration",
		Aggregation: view.LastValue(),
	}
	vLastUpdatedHeight = &view.View{
		Name:        "indexslashing/updated-height",
		Measure:     mUpdatedHeight,
		Description: "Last updated height",
		Aggregation: view.LastValue(),
	}

	views = []*view.View{vRefreshDuration, vLastUpdatedHeight, vRefreshProgress}
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
