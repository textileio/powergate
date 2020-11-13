package tests

import (
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// TxMapDatastore is a in-memory datastore that satisfies TxnDatastore.
type TxMapDatastore struct {
	*datastore.MapDatastore
	lock sync.RWMutex
}

// NewTxMapDatastore returns a new TxMapDatastore.
func NewTxMapDatastore() *TxMapDatastore {
	return &TxMapDatastore{
		MapDatastore: datastore.NewMapDatastore(),
	}
}

// Get returns the value for a key.
func (d *TxMapDatastore) Get(key datastore.Key) ([]byte, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.MapDatastore.Get(key)
}

// Put sets the value of a key.
func (d *TxMapDatastore) Put(key datastore.Key, data []byte) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.MapDatastore.Put(key, data)
}

// Delete deletes a key.
func (d *TxMapDatastore) Delete(key datastore.Key) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.MapDatastore.Delete(key)
}

// Query executes a query in the datastore.
func (d *TxMapDatastore) Query(q query.Query) (query.Results, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.MapDatastore.Query(q)
}

// Clone returns a cloned datastore.
func (d *TxMapDatastore) Clone() (*TxMapDatastore, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	q := query.Query{}
	res, err := d.MapDatastore.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying datastore: %s", err)
	}
	defer func() { _ = res.Close() }()

	t2 := &TxMapDatastore{
		MapDatastore: datastore.NewMapDatastore(),
	}
	for v := range res.Next() {
		if v.Error != nil {
			return nil, fmt.Errorf("iter next: %s", v.Error)
		}
		if err := t2.Put(datastore.NewKey(v.Key), v.Value); err != nil {
			return nil, fmt.Errorf("copying datastore value: %s", err)
		}
	}
	return t2, nil
}

// NewTransaction creates a transaction A read-only transaction should be
// indicated with readOnly equal true.
func (d *TxMapDatastore) NewTransaction(readOnly bool) (datastore.Txn, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return NewSimpleTx(d), nil
}

type op struct {
	delete bool
	value  []byte
}

// SimpleTx implements the transaction interface for datastores who do
// not have any sort of underlying transactional support.
type SimpleTx struct {
	ops    map[datastore.Key]op
	lock   sync.RWMutex
	target datastore.Datastore
}

// NewSimpleTx creates a transaction.
func NewSimpleTx(ds datastore.Datastore) datastore.Txn {
	return &SimpleTx{
		ops:    make(map[datastore.Key]op),
		target: ds,
	}
}

// Query executes a query within the transaction scope.
func (bt *SimpleTx) Query(q query.Query) (query.Results, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Query(q)
}

// Get returns a key value within the transaction.
func (bt *SimpleTx) Get(k datastore.Key) ([]byte, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Get(k)
}

// Has returns true if the key exist, false otherwise.
func (bt *SimpleTx) Has(k datastore.Key) (bool, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Has(k)
}

// GetSize returns the size of the key value.
func (bt *SimpleTx) GetSize(k datastore.Key) (int, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.GetSize(k)
}

// Put sets the value for a key.
func (bt *SimpleTx) Put(key datastore.Key, val []byte) error {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	bt.ops[key] = op{value: val}
	return nil
}

// Delete deletes a key.
func (bt *SimpleTx) Delete(key datastore.Key) error {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	bt.ops[key] = op{delete: true}
	return nil
}

// Discard cancels the changes done in the transaction.
func (bt *SimpleTx) Discard() {
	bt.lock.Lock()
	defer bt.lock.Unlock()
}

// Commit confirms changes done in the transaction.
func (bt *SimpleTx) Commit() error {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	var err error
	for k, op := range bt.ops {
		if op.delete {
			err = bt.target.Delete(k)
		} else {
			err = bt.target.Put(k, op.value)
		}
		if err != nil {
			break
		}
	}

	return err
}
