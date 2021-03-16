package store

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	attrTypeStorage   = attribute.Key("type").String("storage")
	attrTypeRetrieval = attribute.Key("type").String("retrieval")

	attrSuccess = attribute.Key("status").String("success")
	attrFailed  = attribute.Key("status").String("failed")
)

func (s *Store) initMetrics() {
	meter := global.Meter("powergate")

	s.metricFinalTotal = metric.Must(meter).NewInt64Counter("powergate.deals.record.final.total")
	s.metricVolumeBytes = metric.Must(meter).NewInt64Counter("powergate.deals.record.volume.bytes")
}
