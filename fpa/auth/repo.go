package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	fa "github.com/textileio/fil-tools/fpa/fastapi"
)

var (
	ErrNotFound = errors.New("auth token not found")

	dsBase = ds.NewKey("auth")
	log    = logging.Logger("fpaauth")
)

type Repo struct {
	lock sync.Mutex
	ds   ds.Datastore
}

type entry struct {
	Token      string
	InstanceID fa.ID
	// This can be extended to have permissions
}

func New(store ds.Datastore) *Repo {
	return &Repo{
		ds: store,
	}
}

func (r *Repo) Generate(id fa.ID) (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	e := entry{
		Token:      uuid.New().String(),
		InstanceID: id,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		return "", fmt.Errorf("marshaling new auth token for instance %s: %s", id, err)
	}
	if err := r.ds.Put(makeKey(e.Token), buf); err != nil {
		return "", fmt.Errorf("saving generated token from %s to datastore: %s", id, err)
	}
	return e.Token, nil
}

func (r *Repo) Get(token string) (fa.ID, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	buf, err := r.ds.Get(makeKey(token))
	if err != nil && err == ds.ErrNotFound {
		return fa.EmptyID, ErrNotFound
	}
	if err != nil {
		return fa.EmptyID, fmt.Errorf("getting token %s from datastore: %s", token, err)
	}
	var e entry
	if err := json.Unmarshal(buf, &e); err != nil {
		return fa.EmptyID, fmt.Errorf("unmarshaling %s information from datastore: %s", token, err)
	}
	return e.InstanceID, nil
}

func makeKey(token string) ds.Key {
	return dsBase.ChildString(token)
}
