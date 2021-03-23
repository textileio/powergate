package module

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

func (m *Module) initMetrics() {
	meter := global.Meter("powergate")
	m.metricDealTracking = metric.Must(meter).NewInt64UpDownCounter("powergate.deal.tracking")
}
