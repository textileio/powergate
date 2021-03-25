package lotuswallet

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var (
	tagAutofund = attribute.Key("autofunding")
)

func (m *Module) initMetrics() {
	meter := global.Meter("powergate")

	m.metricCreated = metric.Must(meter).NewInt64Counter("powergate.wallet.created")
	m.metricTransfer = metric.Must(meter).NewInt64ValueRecorder("powergate.wallet.transfer", metric.WithDescription("Transfers in nanoFIL"))
}
