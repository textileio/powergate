package source

import (
	"encoding/json"
	"errors"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

var (
	ErrAlreadyExists     = errors.New("source already exists")
	ErrSourceDoesntExist = errors.New("source doesn't exist")

	baseKey = datastore.NewKey("/reputation/store")
)

type SourceStore struct {
	ds datastore.TxnDatastore
}

func NewStore(ds datastore.TxnDatastore) *SourceStore {
	return &SourceStore{
		ds: ds,
	}
}

func (ss *SourceStore) Add(s Source) error {
	txn, err := ss.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	k := genKey(s.Id)
	ok, err := txn.Has(k)
	if err != nil {
		return err
	}
	if ok {
		return ErrAlreadyExists
	}
	return ss.put(txn, s)
}

func (ss *SourceStore) Update(s Source) error {
	txn, err := ss.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	k := genKey(s.Id)
	ok, err := txn.Has(k)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSourceDoesntExist
	}
	return ss.put(txn, s)
}

func (ss *SourceStore) GetAll() ([]Source, error) {
	txn, err := ss.ds.NewTransaction(true)
	if err != nil {
		return nil, err
	}
	q := query.Query{Prefix: baseKey.String()}
	res, err := txn.Query(q)
	if err != nil {
		return nil, err
	}
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

func (ss *SourceStore) put(txn datastore.Txn, s Source) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := txn.Put(genKey(s.Id), b); err != nil {
		return err
	}
	return txn.Commit()

}

func genKey(id string) datastore.Key {
	return baseKey.ChildString(id)
}
