package runner

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/unit"
)

// Init register all views.
func (ai *Runner) initMetrics() error {
	meter := global.Meter("powergate")

	_ = metric.Must(meter).NewFloat64ValueObserver("powergate.index.ask.progress", ai.progressValueObserver, metric.WithDescription("Ask index refresh progress"), metric.WithUnit(unit.Dimensionless))
	ai.refreshDuration = metric.Must(meter).NewInt64ValueRecorder("powergate.index.ask.refresh.duration", metric.WithDescription("Refresh duration"), metric.WithUnit(unit.Milliseconds))

	return nil
}

func (ai *Runner) progressValueObserver(ctx context.Context, result metric.Float64ObserverResult) {
	ai.metricLock.Lock()
	defer ai.metricLock.Unlock()

	result.Observe(ai.progress)
}
