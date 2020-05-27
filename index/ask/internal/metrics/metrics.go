package metrics

import (
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	// TagAskStatus is a tag for query-ask results.
	TagAskStatus, _ = tag.NewKey("askstatus")

	// MFullRefreshDuration is a metric for registering the total duration of a full
	// query ask refresh.
	MFullRefreshDuration = stats.Int64("askindex/fullrefresh-duration", "Duration of Full StorageAsk refresh", "s")
	// MFullRefreshProgress is a metric for registering the current progress of
	// the full scan query ask on miners.
	MFullRefreshProgress = stats.Float64("askindex/fullrefresh-progress", "Full refresh progress", "By")
	// MAskQueryResult is a metric to register the number of results per TagAskStatus status.
	MAskQueryResult = stats.Int64("askindex/queryask-result", "Ask query results", "By")

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

// Init register all views.
func Init() error {
	if err := view.Register(views...); err != nil {
		return fmt.Errorf("register metric views: %v", err)
	}
	return nil
}
