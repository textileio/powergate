package noopauditor

import (
	"context"

	"github.com/google/uuid"
	"github.com/textileio/fil-tools/fpa/types"
)

type Auditor struct {
}

func (np *Auditor) Start(ctx context.Context, instanceID string) types.OpAuditor {
	return &NoopOpAuditor{}
}

type NoopOpAuditor struct {
}

func (noa *NoopOpAuditor) ID() string {
	return uuid.New().String()
}

func (noa *NoopOpAuditor) Success() {

}
func (noa *NoopOpAuditor) Errored(err error) {

}
func (noa *NoopOpAuditor) Log(event interface{}) {

}
func (noa *NoopOpAuditor) Close() {

}
