package store

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/fil-tools/fpa"
)

func (cs *ConfigStore) GetCidInfo(c cid.Cid) (fpa.CidInfo, bool, error) {
	var ci fpa.CidInfo
	buf, err := cs.ds.Get(makeCidInfoKey(cs.iid, c))
	if err == datastore.ErrNotFound {
		return ci, false, nil
	}
	if err != nil {
		return ci, false, err
	}
	if err := json.Unmarshal(buf, &ci); err != nil {
		return ci, false, err
	}
	return ci, true, err
}

func (cs *ConfigStore) SaveCidInfo(cinfo fpa.CidInfo) error {
	buf, err := json.Marshal(cinfo)
	if err != nil {
		return fmt.Errorf("marshaling cidinfo: %s", err)
	}
	if err := cs.ds.Put(makeCidInfoKey(cs.iid, cinfo.Cid), buf); err != nil {
		return fmt.Errorf("saving cidinfo in store: %s", err)
	}
	return nil
}

func (cs *ConfigStore) Cids() ([]cid.Cid, error) {
	q := query.Query{
		Prefix:   makeInstanceKey(cs.iid).String(),
		KeysOnly: true,
	}
	res, err := cs.ds.Query(q)
	if err != nil {
		return nil, err
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
	return makeInstanceKey(iid).ChildString(c.String())
}

func makeInstanceKey(iid fpa.InstanceID) datastore.Key {
	return dsBase.ChildString(iid.String()).Child(dsBaseCidInfo)
}
