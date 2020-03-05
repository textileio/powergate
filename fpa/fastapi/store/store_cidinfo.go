package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/fil-tools/fpa"
	"github.com/textileio/fil-tools/fpa/fastapi"
)

// GetCid info gets the current stored state of a Cid
func (cs *ConfigStore) GetCidInfo(c cid.Cid) (fpa.CidInfo, error) {
	var ci fpa.CidInfo
	buf, err := cs.ds.Get(makeCidInfoKey(cs.iid, c))
	if err == datastore.ErrNotFound {
		return ci, fastapi.ErrCidInfoNotFound
	}
	if err != nil {
		return ci, fmt.Errorf("getting cidinfo %s from store: %s", c, err)
	}
	if err := json.Unmarshal(buf, &ci); err != nil {
		return ci, fmt.Errorf("unmarshaling cidinfo %s: %s", c, err)
	}
	return ci, nil
}

// SaveCidInfo saves a new storing state for a Cid
func (cs *ConfigStore) SaveCidInfo(cinfo fpa.CidInfo) error {
	if !cinfo.Cid.Defined() {
		return fmt.Errorf("cid can't be undefined")
	}
	buf, err := json.Marshal(cinfo)
	if err != nil {
		return fmt.Errorf("marshaling cidinfo: %s", err)
	}
	if err := cs.ds.Put(makeCidInfoKey(cs.iid, cinfo.Cid), buf); err != nil {
		return fmt.Errorf("saving cidinfo in store: %s", err)
	}
	return nil
}

// Cids returns a slice of Cids which have sotring state
func (cs *ConfigStore) Cids() ([]cid.Cid, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	q := query.Query{
		Prefix:   makeInstanceKey(cs.iid).Child(dsBaseCidInfo).String(),
		KeysOnly: true,
	}
	res, err := cs.ds.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying for cids: %s", err)
	}
	defer res.Close()

	var cids []cid.Cid
	for r := range res.Next() {
		strCid := datastore.RawKey(r.Key).Name()
		c, err := cid.Decode(strCid)
		if err != nil {
			return nil, fmt.Errorf("decoding cid: %s", err)
		}
		cids = append(cids, c)
	}
	return cids, nil
}

func makeCidInfoKey(iid fpa.InstanceID, c cid.Cid) datastore.Key {
	return makeInstanceKey(iid).Child(dsBaseCidInfo).ChildString(c.String())
}
