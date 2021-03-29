package dealwatcher

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	attrDealTracked   = attribute.Key("tracked").String("yes")
	attrDealUntracked = attribute.Key("tracked").String("no")
)

func (dw *DealWatcher) initMetrics() {
	meter := global.Meter("powergate")

	dw.metricDealUpdates = metric.Must(meter).NewInt64Counter("powergate.dealwatcher.updates.total")
	dw.metricDealUpdatesChanFailure = metric.Must(meter).NewInt64Counter("powergate.dealwatcher.updates.chan.failure.total")
}
