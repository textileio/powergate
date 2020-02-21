package jsonauditer

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

type JSONAuditer struct {
	ds datastore.Datastore
}

func New(ds datastore.Datastore) *JSONAuditer {
	return &JSONAuditer{
		ds: ds,
	}
}

func (ja *JSONAuditer) Start(ctx context.Context, instanceID string) types.OpAuditer {
	opID := uuid.New().String()
	opa := &JSONOpAuditer{
		id: opID,
		ds: namespace.Wrap(ja.ds, auditDsKey(instanceID, opID)),
	}
	return opa
}

func auditDsKey(iID string, opID string) datastore.Key {
	return dsInstanceBaseKey.ChildString(iID).ChildString(opID)
}

type JSONOpAuditer struct {
	id string
	ds datastore.Datastore
}

func (joa *JSONOpAuditer) ID() string {
	return joa.id
}

func (joa *JSONOpAuditer) Success() {
	panic("TODO")
}
func (joa *JSONOpAuditer) Errored(err error) {
	panic("TODO")
}
func (joa *JSONOpAuditer) Log(event interface{}) {
	panic("TODO")
}
func (joa *JSONOpAuditer) Close() {
	panic("TODO")
}

var _ types.Auditer = (*JSONAuditer)(nil)
var _ types.OpAuditer = (*JSONOpAuditer)(nil)
