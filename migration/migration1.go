package migration

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

var multitenancyMigration = func(txn datastore.Txn, force bool) error {

	if err := migrateJobLogger(txn, force); err != nil {
		return fmt.Errorf("migrating job logger: %s", err)
	}

	if err := migrateStorageInfo(txn, force); err != nil {
		return fmt.Errorf("migrating storage info: %s", err)
	}

	if err := migrateTrackstore(txn, force); err != nil {
		return fmt.Errorf("migrating trackstore: %s", err)
	}

	return nil
}

func migrateJobLogger(txn datastore.Txn, force bool) error {
	panic("TODO")
}

func migrateStorageInfo(txn datastore.Txn, force bool) error {
	panic("TODO")
}

func migrateTrackstore(txn datastore.Txn, force bool) error {
	panic("TODO")
}

func v0_CidOwners(txn datastore.Txn, c cid.Cid) (map[cid.Cid][]ffs.APIID, error) {
	iids, err := v0_APIIDs(txn)
	if err != nil {
		return nil, fmt.Errorf("getting v0 iids: %s", err)
	}

	var owners map[cid.Cid][]ffs.APIID
	for _, iid := range iids {
		cids, err := v0_GetCidsFromIID(txn, iid)
		if err != nil {
			return nil, fmt.Errorf("getting cids from iid: %s", err)
		}
		for _, c := range cids {
			owners[c] = append(owners[c], iid)
		}
	}

	return owners, nil
}

func v0_GetCidsFromIID(txn datastore.Txn, iid ffs.APIID) ([]cid.Cid, error) {
	q := query.Query{Prefix: "/ffs/manager/api/" + iid.String() + "/cidstorageconfig"}
	res, err := txn.Query(q)
	if err != nil {
		return nil, fmt.Errorf("getting cids from iid: %s", err)
	}
	defer func() {
		_ = res.Close()
	}()

	var ret []cid.Cid
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("query result: %s", r.Error)
		}

		// /ffs/manager/api/<iid>/cidstorageconfig/<cid>
		cidStr := datastore.NewKey(r.Key).Namespaces()[5]
		c, err := util.CidFromString(cidStr)
		if err != nil {
			return nil, fmt.Errorf("discovered invalid cid: %s", err)
		}
		ret = append(ret, c)
	}

	return ret, nil
}

func v0_APIIDs(txn datastore.Txn) ([]ffs.APIID, error) {
	q := query.Query{Prefix: "/ffs/manager/api"}
	res, err := txn.Query(q)
	if err != nil {
		return nil, fmt.Errorf("getting iids: %s", err)
	}
	defer func() {
		_ = res.Close()
	}()

	var ret []ffs.APIID
	for r := range res.Next() {
		if r.Error != nil {
			return nil, fmt.Errorf("query result: %s", r.Error)
		}

		k := datastore.NewKey(r.Key)
		iid := ffs.APIID(k.Namespaces()[3]) // /ffs/manager/api/<iid>/...
		if !iid.Valid() {
			return nil, fmt.Errorf("discovered invalid iid: %s", err)
		}
		ret = append(ret, iid)
	}

	return ret, nil
}
