package source

import (
	"encoding/json"
	"errors"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

var (
	// ErrAlreadyExists returns when the soure already exists in Store
	ErrAlreadyExists = errors.New("source already exists")
	// ErrDoesntExists returns when the source isn't in the Store
	ErrDoesntExists = errors.New("source doesn't exist")

	baseKey = datastore.NewKey("/reputation/store")
)

// Store contains Sources information
type Store struct {
	ds datastore.TxnDatastore
}

// NewStore returns a new SourceStore
func NewStore(ds datastore.TxnDatastore) *Store {
	return &Store{
		ds: ds,
	}
}

// Add adds a new Source to the store
func (ss *Store) Add(s Source) error {
	txn, err := ss.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	k := genKey(s.ID)
	ok, err := txn.Has(k)
	if err != nil {
		return err
	}
	if ok {
		return ErrAlreadyExists
	}
	return ss.put(txn, s)
}

// Update updates a Source
func (ss *Store) Update(s Source) error {
	txn, err := ss.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	k := genKey(s.ID)
	ok, err := txn.Has(k)
	if err != nil {
		return err
	}
	if !ok {
		return ErrDoesntExists
	}
	return ss.put(txn, s)
}

// GetAll returns all Sources
func (ss *Store) GetAll() ([]Source, error) {
	txn, err := ss.ds.NewTransaction(true)
	if err != nil {
		return nil, err
	}
	q := query.Query{Prefix: baseKey.String()}
	res, err := txn.Query(q)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var ret []Source
	for r := range res.Next() {
		s := Source{}
		if err := json.Unmarshal(r.Value, &s); err != nil {
			return nil, err
		}
		ret = append(ret, s)
	}
	return ret, nil
}

func (ss *Store) put(txn datastore.Txn, s Source) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := txn.Put(genKey(s.ID), b); err != nil {
		return err
	}
	return txn.Commit()

}

func genKey(id string) datastore.Key {
	return baseKey.ChildString(id)
}
