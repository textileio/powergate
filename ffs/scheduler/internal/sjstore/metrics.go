package sjstore

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	attrStatusQueued    = attribute.Key("jobstatus").String("queued")
	attrStatusExecuting = attribute.Key("jobstatus").String("executing")
	attrStatusSuccess   = attribute.Key("jobstatus").String("success")
	attrStatusCanceled  = attribute.Key("jobstatus").String("canceled")
	attrStatusFailed    = attribute.Key("jobstatus").String("failed")
)

func (s *Store) initMetrics() {
	meter := global.Meter("powergate")
	s.metricJobCounter = metric.Must(meter).NewInt64UpDownCounter("powergate.storage.job.total")
}
