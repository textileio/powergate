package module

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
//attrStatusQueued    = attribute.Key("jobstatus").String("queued")
)

func (m *Module) initMetrics() {
	meter := global.Meter("powergate")
	m.metricDealTracking = metric.Must(meter).NewInt64UpDownCounter("powergate.deal.tracking")
}
