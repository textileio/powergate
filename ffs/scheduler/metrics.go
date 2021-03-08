package scheduler

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	metricTagJobStatusQueued    = attribute.Key("jobstatus").String("queued")
	metricTagJobStatusExecuting = attribute.Key("jobstatus").String("executing")

	metricTagPreprocessingStatusWaiting    = attribute.Key("status").String("waiting")
	metricTagPreprocessingStatusInProgress = attribute.Key("status").String("inprogress")
)

func (s *Scheduler) initMetrics() {
	meter := global.Meter("powergate")
	s.metricPreprocessingTotal = metric.Must(meter).NewInt64UpDownCounter("Preprocessing job queue")
	_ = metric.Must(meter).NewInt64ValueObserver("Storage deal queue", s.observeStorageDealQueue)
}

func (s *Scheduler) observeStorageDealQueue(ctx context.Context, result metric.Int64ObserverResult) {
	stats := s.sjs.GetStats()
	result.Observe(int64(stats.TotalQueued), metricTagJobStatusQueued)
	result.Observe(int64(stats.TotalExecuting), metricTagJobStatusExecuting)
}
