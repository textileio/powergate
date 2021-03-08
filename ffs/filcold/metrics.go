package filcold

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	metricTagPreprocessingWaiting    = attribute.Key("status").String("waiting")
	metricTagPreprocessingInProgress = attribute.Key("status").String("inprogress")
)

func (fc *FilCold) initMetrics() {
	meter := global.Meter("powergate")
	fc.metricPreprocessingTotal = metric.Must(meter).NewInt64UpDownCounter("powergate.preprocessing.queue.total")
}
