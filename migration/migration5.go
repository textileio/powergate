package migration

import (
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// V5DeleteOldMinerIndex contains the logic to upgrade a datastore from
// version 4 to version 5.
var V5DeleteOldMinerIndex = Migration{
	UseTxn: false,
	Run: func(ds datastoreReaderWriter) error {
		q := query.Query{Prefix: "/index/miner/chainstore"}
		res, err := ds.Query(q)
		if err != nil {
			return fmt.Errorf("querying records: %s", err)
		}
		defer func() { _ = res.Close() }()

		var count int
		for v := range res.Next() {
			if err := ds.Delete(datastore.NewKey(v.Key)); err != nil {
				return fmt.Errorf("deleting miner chainstore key: %s", err)
			}
			count++
		}
		log.Infof("deleted %d chainstore keys", count)

		return nil
	},
}
