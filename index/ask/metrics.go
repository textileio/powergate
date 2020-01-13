package ask

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mAskCacheFullRefreshTime = stats.Int64("askcache/full-refresh-time", "How many seconds take full Ask cache refresh", "s")
	mAskQuery                = stats.Int64("askcache/askquery", "Ask query results", "By")
	mFullRefreshProgress     = stats.Float64("askcache/full-refresh-progress", "Full refresh progress", "By")

	vAskCacheFullRefreshTime = &view.View{
		Name:        "askcache/full-refresh-time",
		Measure:     mAskCacheFullRefreshTime,
		Description: "Ask cache full refresh time",
		Aggregation: view.LastValue(),
	}
	vAskQuery = &view.View{
		Name:        "askcache/askquery",
		Measure:     mAskQuery,
		Description: "Ask query results",
		TagKeys:     []tag.Key{keyAskStatus},
		Aggregation: view.LastValue(),
	}
	vFullRefreshProgress = &view.View{
		Name:        "askcache/full-refresh-progress",
		Measure:     mFullRefreshProgress,
		Description: "Full refresh progress",
		Aggregation: view.LastValue(),
	}
	views = []*view.View{vAskCacheFullRefreshTime, vAskQuery, vFullRefreshProgress}

	keyAskStatus, _ = tag.NewKey("askstatus")
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
