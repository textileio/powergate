package sendstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/wallet"
)

var (
	log = logging.Logger("wallet-sendstore")

	// ErrNotFound indicates the instance doesn't exist.
	ErrNotFound = errors.New("not found")

	dsBaseEvent     = datastore.NewKey("event").String()
	dsBaseIndexFrom = datastore.NewKey("index").ChildString("from").String()
	dsBaseIndexTo   = datastore.NewKey("index").ChildString("to").String()
)

// SendStore stores information about SendFil transactions.
type SendStore struct {
	ds datastore.TxnDatastore
}

// New creates a new SendStore.
func New(ds datastore.TxnDatastore) *SendStore {
	return &SendStore{
		ds: ds,
	}
}

// Put saves a transaction.
func (s *SendStore) Put(cid cid.Cid, from, to address.Address, amount *big.Int) (*wallet.SendFilEvent, error) {
	rec := &wallet.SendFilEvent{
		Cid:    cid,
		From:   from,
		To:     to,
		Amount: amount,
		Time:   time.Now(),
	}
	bytes, err := json.Marshal(rec)
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %v", err)
	}

	tx, err := s.ds.NewTransaction(false)
	defer tx.Discard()
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %v", err)
	}

	dataKey := eventKey(cid)

	err = tx.Put(dataKey, bytes)
	if err != nil {
		return nil, fmt.Errorf("putting rec: %v", err)
	}

	err = tx.Put(indexFromKey(cid, from, to, rec.Time), dataKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("putting from index: %v", err)
	}

	err = tx.Put(indexToKey(cid, to, from, rec.Time), dataKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("putting to index: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("committing transaction: %v", err)
	}
	return rec, nil
}

// Get retrieves a send fil txn by cid.
func (s *SendStore) Get(cid cid.Cid) (*wallet.SendFilEvent, error) {
	return s.get(eventKey(cid))
}

// From returns all SendFilEvents sent from the specified address.
func (s *SendStore) From(address address.Address) ([]*wallet.SendFilEvent, error) {
	return s.withIndexPrefix(indexFromPrefix(address))
}

// To returns all SendFilEvents sent to the specified address.
func (s *SendStore) To(address address.Address) ([]*wallet.SendFilEvent, error) {
	return s.withIndexPrefix(indexToPrefix(address))
}

// FromTo returns all SendFilEvents sent from the specified address to the specified address.
func (s *SendStore) FromTo(from, to address.Address) ([]*wallet.SendFilEvent, error) {
	return s.withIndexPrefix(indexFromToPrefix(from, to))
}

// Between returns all SendFilEvents between the specified addresses.
func (s *SendStore) Between(addr1, addr2 address.Address) ([]*wallet.SendFilEvent, error) {
	res1, err := s.withIndexPrefix(indexFromToPrefix(addr1, addr2))
	if err != nil {
		return nil, fmt.Errorf("getting events from addr1 to addr 2: %v", err)
	}
	res2, err := s.withIndexPrefix(indexFromToPrefix(addr2, addr1))
	if err != nil {
		return nil, fmt.Errorf("getting events from addr1 to addr 2: %v", err)
	}
	return append(res1, res2...), nil
}

func (s *SendStore) withIndexPrefix(prefix string) ([]*wallet.SendFilEvent, error) {
	q := query.Query{Prefix: prefix}
	res, err := s.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying datastore: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Errorf("closing allWithIndexPrefix index query result: %s", err)
		}
	}()
	var events []*wallet.SendFilEvent
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("iter next: %s", r.Error)
		}
		eventKey := datastore.NewKey(string(r.Value))
		event, err := s.get(eventKey)
		if err != nil {
			return nil, fmt.Errorf("getting event: %v", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (s *SendStore) get(key datastore.Key) (*wallet.SendFilEvent, error) {
	bytes, err := s.ds.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getting event bytes from ds: %v", err)
	}
	event := &wallet.SendFilEvent{}
	err = json.Unmarshal(bytes, event)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling bytes into event: %v", err)
	}
	return event, nil
}

func eventKey(cid cid.Cid) datastore.Key {
	return datastore.NewKey(dsBaseEvent).ChildString(cid.String())
}

func indexFromPrefix(from address.Address) string {
	return fmt.Sprintf("%s/%s", dsBaseIndexFrom, from.String())
}

func indexToPrefix(to address.Address) string {
	return fmt.Sprintf("%s/%s", dsBaseIndexTo, to.String())
}

func indexFromToPrefix(from, to address.Address) string {
	return fmt.Sprintf("%s/%s", indexFromPrefix(from), to.String())
}

func indexToFromPrefix(to, from address.Address) string {
	return fmt.Sprintf("%s/%s", indexToPrefix(to), from.String())
}

func indexFromKey(cid cid.Cid, from, to address.Address, time time.Time) datastore.Key {
	return datastore.NewKey(indexFromToPrefix(from, to)).ChildString(fmt.Sprintf("%d", time.UnixNano())).ChildString(cid.String())
}

func indexToKey(cid cid.Cid, to, from address.Address, time time.Time) datastore.Key {
	return datastore.NewKey(indexToFromPrefix(to, from)).ChildString(fmt.Sprintf("%d", time.UnixNano())).ChildString(cid.String())
}
