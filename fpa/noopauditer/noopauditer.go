package noopauditer

import (
	"context"

	"github.com/google/uuid"
	"github.com/textileio/fil-tools/fpa/types"
)

type Auditer struct {
}

func (np *Auditer) Start(ctx context.Context, instanceID string) types.OpAuditer {
	return &NoopOpAuditer{}
}

type NoopOpAuditer struct {
}

func (noa *NoopOpAuditer) ID() string {
	return uuid.New().String()
}

func (noa *NoopOpAuditer) Success() {

}
func (noa *NoopOpAuditer) Errored(err error) {

}
func (noa *NoopOpAuditer) Log(event interface{}) {

}
func (noa *NoopOpAuditer) Close() {

}
