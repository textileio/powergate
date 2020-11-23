package migration

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

// V1MultitenancyMigration contains the logic to upgrade a datastore from
// version 0 to version 1.
var V1MultitenancyMigration = Migration(func(txn datastore.Txn) error {
	cidOwners, err := v0CidOwners(txn)
	if err != nil {
		return fmt.Errorf("getting cid owners: %s", err)
	}

	log.Infof("Starting job logger migration...")
	if err := migrateJobLogger(txn, cidOwners); err != nil {
		return fmt.Errorf("migrating job logger: %s", err)
	}
	log.Infof("Job logger migration finished")

	log.Infof("Starting storage info migration...")
	if err := migrateStorageInfo(txn, cidOwners); err != nil {
		return fmt.Errorf("migrating storage info: %s", err)
	}
	log.Infof("Storage info migration finished")

	log.Infof("Starting trackstore migration...")
	if err := migrateTrackstore(txn, cidOwners); err != nil {
		return fmt.Errorf("migrating trackstore: %s", err)
	}
	log.Infof("Trackstore migration finished")

	return nil
})

func v0CidOwners(txn datastore.Txn) (map[cid.Cid][]ffs.APIID, error) {
	iids, err := v0APIIDs(txn)
	if err != nil {
		return nil, fmt.Errorf("getting v0 iids: %s", err)
	}

	owners := map[cid.Cid][]ffs.APIID{}
	for _, iid := range iids {
		cids, err := v0GetCidsFromIID(txn, iid)
		if err != nil {
			return nil, fmt.Errorf("getting cids from iid: %s", err)
		}
		for _, c := range cids {
			owners[c] = append(owners[c], iid)
		}
	}

	return owners, nil
}

func v0GetCidsFromIID(txn datastore.Txn, iid ffs.APIID) ([]cid.Cid, error) {
	q := query.Query{Prefix: "/ffs/manager/api/" + iid.String() + "/istore/cidstorageconfig"}
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

		// /ffs/manager/api/<iid>/istore/cidstorageconfig/<cid>
		cidStr := datastore.NewKey(r.Key).Namespaces()[6]
		c, err := util.CidFromString(cidStr)
		if err != nil {
			return nil, fmt.Errorf("discovered invalid cid: %s", err)
		}
		ret = append(ret, c)
	}

	return ret, nil
}

func v0APIIDs(txn datastore.Txn) ([]ffs.APIID, error) {
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

func v0GetStorageConfig(txn datastore.Txn, iid ffs.APIID, c cid.Cid) (ffs.StorageConfig, error) {
	if !iid.Valid() {
		return ffs.StorageConfig{}, fmt.Errorf("invalid iid %s", iid)
	}
	if !c.Defined() {
		return ffs.StorageConfig{}, fmt.Errorf("undefined cid")
	}
	key := datastore.NewKey("/ffs/manager/api/" + iid.String() + "/istore/cidstorageconfig/" + util.CidToString(c))

	buf, err := txn.Get(key)
	if err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("getting storage config: %s", err)
	}
	var conf ffs.StorageConfig
	if err := json.Unmarshal(buf, &conf); err != nil {
		return ffs.StorageConfig{}, fmt.Errorf("unmarshaling cid config from datastore: %s", err)
	}

	return conf, nil
}
