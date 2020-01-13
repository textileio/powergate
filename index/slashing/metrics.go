package slashing

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	mRefreshProgress = stats.Float64("index/slashing-refresh-progress", "Refresh progress", "By")
	mRefreshDuration = stats.Int64("index/slashing-full-refresh-duration", "Seconds of full-refresh", "s")
	mUpdatedHeight   = stats.Int64("index/slashing-updated-height", "Last updated height", "By")

	vRefreshProgress = &view.View{
		Name:        "index/slashing-refresh-progress",
		Measure:     mRefreshProgress,
		Description: "Refresh duration",
		Aggregation: view.LastValue(),
	}
	vRefreshDuration = &view.View{
		Name:        "index/slashing-refresh-duration",
		Measure:     mRefreshDuration,
		Description: "Refresh duration",
		Aggregation: view.LastValue(),
	}
	vLastUpdatedHeight = &view.View{
		Name:        "index/slashing-updated-height",
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
