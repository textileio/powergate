package lotusidx

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mOnChainRefreshProgress = stats.Float64("indexminer/onchain-refresh-progress", "On-chain refresh progress", "By")
	mMetaRefreshProgress    = stats.Float64("indexminer/meta-refresh-progress", "Meta refresh progress", "By")
	mRefreshDuration        = stats.Int64("indexminer/refresh-duration", "Refresh duration in seconds", "s")
	mMetaPingCount          = stats.Int64("indexminer/meta-ping-count", "On-chain miners ping response", "By")
	mUpdatedHeight          = stats.Int64("indexminer/updated-height", "Last updated height", "By")

	vOnChainRefreshTimeProgress = &view.View{
		Name:        "indexminer/onchain-refresh-progress",
		Measure:     mOnChainRefreshProgress,
		Description: "OnChain refresh progress",
		Aggregation: view.LastValue(),
	}
	vMetaRefreshTimeProgress = &view.View{
		Name:        "indexminer/meta-refresh-progress",
		Measure:     mMetaRefreshProgress,
		Description: "Meta refresh progress",
		Aggregation: view.LastValue(),
	}
	vRefreshDuration = &view.View{
		Name:        "indexminer/refresh-duration",
		Measure:     mRefreshDuration,
		Description: "Refresh duration",
		TagKeys:     []tag.Key{metricRefreshType},
		Aggregation: view.LastValue(),
	}
	vMetaPingCount = &view.View{
		Name:        "indexminer/meta-ping-count",
		Measure:     mMetaPingCount,
		Description: "On-chain miners ping response",
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

	views = []*view.View{vOnChainRefreshTimeProgress, vRefreshDuration, vMetaPingCount, vUpdatedHeight, vMetaRefreshTimeProgress}
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
