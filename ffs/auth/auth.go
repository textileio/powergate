package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/v2/ffs"
)

var (
	// ErrNotFound indicates that the auth-token isn't registered.
	ErrNotFound = errors.New("auth token not found")

	log = logging.Logger("ffs-auth")
)

// Auth contains a mapping between auth-tokens and Api instances.
type Auth struct {
	lock sync.Mutex
	ds   ds.Datastore
}

// New returns a new Auth.
func New(store ds.Datastore) *Auth {
	return &Auth{
		ds: store,
	}
}

// Generate generates a new returned auth-token mapped to the iid.
func (r *Auth) Generate(iid ffs.APIID) (string, error) {
	log.Infof("generating auth-token for instance %s", iid)
	r.lock.Lock()
	defer r.lock.Unlock()
	e := ffs.AuthEntry{
		Token: uuid.New().String(),
		APIID: iid,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		return "", fmt.Errorf("marshaling new auth token for instance %s: %s", iid, err)
	}
	if err := r.ds.Put(ds.NewKey(e.Token), buf); err != nil {
		return "", fmt.Errorf("saving generated token from %s to datastore: %s", iid, err)
	}
	return e.Token, nil
}

// Get returns the InstanceID associated with token.
// It returns ErrNotFound if there isn't such.
func (r *Auth) Get(token string) (ffs.APIID, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	buf, err := r.ds.Get(ds.NewKey(token))
	if err != nil && err == ds.ErrNotFound {
		return ffs.EmptyInstanceID, ErrNotFound
	}
	if err != nil {
		return ffs.EmptyInstanceID, fmt.Errorf("getting token %s from datastore: %s", token, err)
	}
	var e ffs.AuthEntry
	if err := json.Unmarshal(buf, &e); err != nil {
		return ffs.EmptyInstanceID, fmt.Errorf("unmarshaling %s information from datastore: %s", token, err)
	}
	return e.APIID, nil
}

// List returns a list of all API instances.
func (r *Auth) List() ([]ffs.AuthEntry, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	q := query.Query{Prefix: ""}
	res, err := r.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("executing query in datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing query result: %s", err)
		}
	}()

	var ret []ffs.AuthEntry
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		var e ffs.AuthEntry
		if err := json.Unmarshal(r.Entry.Value, &e); err != nil {
			return nil, fmt.Errorf("unmarshaling query result: %s", err)
		}
		ret = append(ret, e)
	}
	return ret, nil
}
