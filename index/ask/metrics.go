package ask

import (
	"github.com/prometheus/common/log"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	mAskCacheFullRefreshTime = stats.Int64("askcache/full-refresh-time", "How many seconds take full Ask cache refresh", "s")
	mMinerCount              = stats.Int64("askcache/mes-miner-count", "How many miners onchain", "By")
	mAskQuery                = stats.Int64("askcache/askquery", "Ask query results", "By")

	viewAskCacheFullRefreshTime = &view.View{
		Name:        "askcache/full-refresh-time",
		Measure:     mAskCacheFullRefreshTime,
		Description: "Ask cache full refresh time",
		Aggregation: view.LastValue(),
	}
	viewMinerCount = &view.View{
		Name:        "askcache/miner_count",
		Measure:     mMinerCount,
		Description: "Miners on chain",
		Aggregation: view.LastValue(),
	}
	viewAskQuery = &view.View{
		Name:        "askcache/askquery",
		Measure:     mAskQuery,
		Description: "Ask query results",
		TagKeys:     []tag.Key{keyAskStatus},
		Aggregation: view.LastValue(),
	}
	views = []*view.View{viewAskCacheFullRefreshTime, viewMinerCount, viewAskQuery}

	keyAskStatus, _ = tag.NewKey("askstatus")
)

func initMetrics() {
	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
}
