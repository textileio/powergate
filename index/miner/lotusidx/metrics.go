package lotusidx

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/unit"
)

var (
	onchainSubindex = attribute.Key("subindex").String("onchain")
	metaSubindex    = attribute.Key("subindex").String("meta")
)

func (mi *Index) initMetrics() {
	meter := global.Meter("powergate")

	_ = metric.Must(meter).NewFloat64ValueObserver("powergate.index.miner.progress", mi.progressValueObserver, metric.WithDescription("Miner index refresh progress"), metric.WithUnit(unit.Dimensionless))
	mi.meterRefreshDuration = metric.Must(meter).NewInt64ValueRecorder("powergate.index.miner.refresh.duration", metric.WithDescription("Refresh duration"), metric.WithUnit(unit.Milliseconds))
}

func (mi *Index) progressValueObserver(ctx context.Context, result metric.Float64ObserverResult) {
	mi.metricLock.Lock()
	defer mi.metricLock.Unlock()
	result.Observe(mi.onchainProgress, onchainSubindex)
	result.Observe(mi.metaProgress, metaSubindex)
}
