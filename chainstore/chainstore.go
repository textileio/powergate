package chainstore

import (
	"context"
	"fmt"
	"math/bits"
	"sort"
	"strconv"
	"sync"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	cbor "github.com/ipfs/go-ipld-cbor"
)

const (
	maxCheckpoints = 10
)

var (
	dsNsData = datastore.NewKey("/data")
	dsNsID   = datastore.NewKey("/id")
)

// TipsetOrderer resolves ordering information between TipSets
type TipsetOrderer interface {
	Precedes(ctx context.Context, from, to types.TipSetKey) (bool, error)
}

type checkpoint struct {
	id uint64
	ts types.TipSetKey
}

// Store allows to save snapshoted state
type Store struct {
	fr TipsetOrderer

	lock        sync.Mutex
	ds          datastore.TxnDatastore
	checkpoints []checkpoint
	lastID      uint64
}

// New returns a new Store
func New(ds datastore.TxnDatastore, fr TipsetOrderer) (*Store, error) {
	s := &Store{
		ds: ds,
		fr: fr,
	}
	if err := s.loadCheckpoints(); err != nil {
		return nil, err
	}
	return s, nil
}

// GetLastCheckpoint returns the most recent saved state, and its TipSetKey.
// In case none exist, it will return a nil TipSetKey.
func (s *Store) GetLastCheckpoint(v interface{}) (*types.TipSetKey, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.checkpoints) == 0 {
		return nil, nil
	}
	chk := s.checkpoints[len(s.checkpoints)-1]
	if err := s.load(chk.ts, v); err != nil {
		return nil, err
	}
	return &chk.ts, nil
}

// LoadAndPrune loads into value the last saved state before the currHead tipset branch,
// and returns the TipSetKey of that state. If currHead invalidates saved checkpoints
// those will be deleted (auto-pruning).
func (s *Store) LoadAndPrune(ctx context.Context, currHead types.TipSetKey, value interface{}) (*types.TipSetKey, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var base *types.TipSetKey
	for i := len(s.checkpoints) - 1; i >= 0 && base == nil; i-- {
		c := s.checkpoints[i]
		ok, err := s.fr.Precedes(ctx, c.ts, currHead)
		if err != nil {
			return nil, err
		}
		if !ok {
			if err := s.delete(c); err != nil {
				return nil, err
			}
			s.checkpoints = s.checkpoints[:i]
			continue
		}
		base = &c.ts
	}
	if base == nil {
		return nil, nil
	}
	if err := s.load(*base, value); err != nil {
		return nil, err
	}
	return base, nil
}

func (s *Store) load(tsk types.TipSetKey, v interface{}) error {
	key := toKeyData(tsk)
	buf, err := s.ds.Get(key)
	if err != nil {
		return fmt.Errorf("error getting key %s: %s", key, err)
	}
	if err := cbor.DecodeInto(buf, v); err != nil {
		return err
	}
	return nil
}

// Save saves a new current state which should be a child of the last known checkpoint.
func (s *Store) Save(ctx context.Context, ts types.TipSetKey, state interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.checkpoints) > 0 {
		last := s.checkpoints[len(s.checkpoints)-1]
		ok, err := s.fr.Precedes(ctx, last.ts, ts)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("new state should be on the same chain as last checkpoint")
		}
	}
	if err := s.save(ts, state); err != nil {
		return err
	}
	return nil
}

func (s *Store) save(ts types.TipSetKey, state interface{}) error {
	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return nil
	}
	defer txn.Discard()
	buf, err := cbor.DumpObject(state)
	if err != nil {
		return err
	}
	c := checkpoint{id: s.lastID + 1, ts: ts}
	if err := txn.Put(toKeyData(c.ts), buf); err != nil {
		return err
	}
	if err := txn.Put(toKeyID(c.id), c.ts.Bytes()); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}

	s.checkpoints = append(s.checkpoints, c)
	if len(s.checkpoints) > maxCheckpoints {
		dc := s.checkpoints[0]
		if err := s.delete(dc); err != nil {
			return err
		}
		copy(s.checkpoints, s.checkpoints[1:])
		s.checkpoints = s.checkpoints[:len(s.checkpoints)-1]
	}
	s.lastID++

	return nil
}

func (s *Store) delete(c checkpoint) error {
	txn, err := s.ds.NewTransaction(false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	if err := txn.Delete(toKeyData(c.ts)); err != nil {
		return err
	}
	if err := txn.Delete(toKeyID(c.id)); err != nil {
		return err
	}
	return txn.Commit()
}

func toKeyData(ts types.TipSetKey) datastore.Key {
	return dsNsData.ChildString(string(ts.Bytes()))
}

func toKeyID(id uint64) datastore.Key {
	return dsNsID.ChildString(strconv.FormatUint(id, 10))
}

func (s *Store) loadCheckpoints() error {
	res, err := s.ds.Query(query.Query{Prefix: dsNsID.String()})
	if err != nil {
		return err
	}
	defer res.Close()
	es, err := res.Rest()
	if err != nil {
		return err
	}
	lst := make([]checkpoint, len(es))
	for i, e := range es {
		ts, err := types.TipSetKeyFromBytes([]byte(e.Value))
		if err != nil {
			return err
		}
		id, err := strconv.ParseUint(datastore.RawKey(e.Key).List()[1], 10, bits.UintSize)
		if err != nil {
			return err
		}
		lst[i] = checkpoint{
			id: id,
			ts: ts,
		}
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].id < lst[j].id
	})
	if len(lst) > 0 {
		s.lastID = lst[len(lst)-1].id
	}
	s.checkpoints = lst
	return nil
}
