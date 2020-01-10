package miner

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mRefreshProgress = stats.Float64("index/miner-refresh-progress", "Refresh progress", "By")
	mRefreshTime     = stats.Int64("index/miner-refresh-time", "Seconds of full-refresh", "s")
	mMinerCount      = stats.Int64("index/miner-count", "How many miners on-chain", "By")
	mUpdatedHeight   = stats.Int64("index/miner-updated-height", "Last updated height", "By")

	vRefreshTimeProgress = &view.View{
		Name:        "index/miner-refresh-progress",
		Measure:     mRefreshProgress,
		Description: "Refresh time",
		Aggregation: view.LastValue(),
	}
	vRefreshTime = &view.View{
		Name:        "index/miner-refresh-time",
		Measure:     mRefreshTime,
		Description: "Refresh time",
		TagKeys:     []tag.Key{kRefreshType},
		Aggregation: view.LastValue(),
	}
	vMinerCount = &view.View{
		Name:        "index/miner-count",
		Measure:     mMinerCount,
		Description: "Miners on chain",
		TagKeys:     []tag.Key{kOnline},
		Aggregation: view.LastValue(),
	}
	vLastUpdatedHeight = &view.View{
		Name:        "index/miner-updated-height",
		Measure:     mUpdatedHeight,
		Description: "Last updated height",
		Aggregation: view.LastValue(),
	}
	kRefreshType, _ = tag.NewKey("refreshtype")
	kOnline, _      = tag.NewKey("online")

	views = []*view.View{vRefreshTime, vMinerCount, vLastUpdatedHeight, vRefreshTimeProgress}
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
