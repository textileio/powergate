package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
)

var (
	// ErrNotFound indicates that the auth-token isn't registered
	ErrNotFound = errors.New("auth token not found")

	dsBase = ds.NewKey("auth")
	log    = logging.Logger("ffs-auth")
)

// Auth contains a mapping between auth-tokens and Api instances.
type Auth struct {
	lock sync.Mutex
	ds   ds.Datastore
}

type entry struct {
	Token      string
	InstanceID ffs.ApiID
	// This can be extended to have permissions
}

// New returns a new Auth
func New(store ds.Datastore) *Auth {
	return &Auth{
		ds: store,
	}
}

// Generate generates a new returned auth-token mapped to the iid
func (r *Auth) Generate(iid ffs.ApiID) (string, error) {
	log.Infof("generating auth-token for instance %s", iid)
	r.lock.Lock()
	defer r.lock.Unlock()
	e := entry{
		Token:      uuid.New().String(),
		InstanceID: iid,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		return "", fmt.Errorf("marshaling new auth token for instance %s: %s", iid, err)
	}
	if err := r.ds.Put(makeKey(e.Token), buf); err != nil {
		return "", fmt.Errorf("saving generated token from %s to datastore: %s", iid, err)
	}
	return e.Token, nil
}

// Get returns the InstanceID associated with token.
// It returns ErrNotFound if there isn't such.
func (r *Auth) Get(token string) (ffs.ApiID, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	buf, err := r.ds.Get(makeKey(token))
	if err != nil && err == ds.ErrNotFound {
		return ffs.EmptyInstanceID, ErrNotFound
	}
	if err != nil {
		return ffs.EmptyInstanceID, fmt.Errorf("getting token %s from datastore: %s", token, err)
	}
	var e entry
	if err := json.Unmarshal(buf, &e); err != nil {
		return ffs.EmptyInstanceID, fmt.Errorf("unmarshaling %s information from datastore: %s", token, err)
	}
	return e.InstanceID, nil
}

func makeKey(token string) ds.Key {
	return dsBase.ChildString(token)
}
