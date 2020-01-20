package miner

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mRefreshProgress   = stats.Float64("indexminer/refresh-progress", "Refresh progress", "By")
	mRefreshDuration   = stats.Int64("indexminer/refresh-duration", "Refresh duration in seconds", "s")
	mMinerOnChainCount = stats.Int64("indexminer/onchain-count", "How many miners on-chain", "By")
	mUpdatedHeight     = stats.Int64("indexminer/updated-height", "Last updated height", "By")

	vRefreshTimeProgress = &view.View{
		Name:        "indexminer/refresh-progress",
		Measure:     mRefreshProgress,
		Description: "Refresh progress",
		Aggregation: view.LastValue(),
	}
	vRefreshDuration = &view.View{
		Name:        "indexminer/refresh-duration",
		Measure:     mRefreshDuration,
		Description: "Refresh duration",
		TagKeys:     []tag.Key{metricRefreshType},
		Aggregation: view.LastValue(),
	}
	vMinerOnChainCount = &view.View{
		Name:        "indexminer/onchain-count",
		Measure:     mMinerOnChainCount,
		Description: "Miners on chain",
		TagKeys:     []tag.Key{metricOnline},
		Aggregation: view.LastValue(),
	}
	vUpdatedHeight = &view.View{
		Name:        "index/miner-updated-height",
		Measure:     mUpdatedHeight,
		Description: "Last updated height",
		Aggregation: view.LastValue(),
	}
	metricRefreshType, _ = tag.NewKey("refreshtype")
	metricOnline, _      = tag.NewKey("online")

	views = []*view.View{vRefreshTimeProgress, vRefreshDuration, vMinerOnChainCount, vUpdatedHeight}
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
