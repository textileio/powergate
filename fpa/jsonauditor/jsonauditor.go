package jsonauditor

import (
	"context"

	"github.com/google/uuid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/textileio/fil-tools/fpa/types"
)

var (
	dsInstanceBaseKey = datastore.NewKey("instance")
)

type JSONAuditor struct {
	ds datastore.Datastore
}

func New(ds datastore.Datastore) *JSONAuditor {
	return &JSONAuditor{
		ds: ds,
	}
}

func (ja *JSONAuditor) Start(ctx context.Context, instanceID string) types.OpAuditor {
	opID := uuid.New().String()
	opa := &JSONOpAuditor{
		id: opID,
		ds: namespace.Wrap(ja.ds, auditDsKey(instanceID, opID)),
	}
	return opa
}

func auditDsKey(iID string, opID string) datastore.Key {
	return dsInstanceBaseKey.ChildString(iID).ChildString(opID)
}

type JSONOpAuditor struct {
	id string
	ds datastore.Datastore
}

func (joa *JSONOpAuditor) ID() string {
	return joa.id
}

func (joa *JSONOpAuditor) Success() {
	panic("TODO")
}
func (joa *JSONOpAuditor) Errored(err error) {
	panic("TODO")
}
func (joa *JSONOpAuditor) Log(event interface{}) {
	panic("TODO")
}
func (joa *JSONOpAuditor) Close() {
	panic("TODO")
}

var _ types.Auditor = (*JSONAuditor)(nil)
var _ types.OpAuditor = (*JSONOpAuditor)(nil)
