package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/textileio/powergate/v2/cmd/powcfg/ratelim"
	"github.com/textileio/powergate/v2/ffs"
)

// Transform is a function that applies a transformation to a StorageConfig.
// The function should edit the received parameter directly, and return 'true'
// if changes were made, or false otherwise.
type Transform func(*ffs.StorageConfig) (bool, error)

func applyTransform(ds datastore.TxnDatastore, dryrun bool, t Transform) (int, error) {
	q := query.Query{Prefix: "/ffs/manager/api"}
	res, err := ds.Query(q)
	if err != nil {
		return 0, fmt.Errorf("running query: %s", err)
	}
	defer func() {
		if err := res.Close(); err != nil {
			log.Warnf("closing query result: %s", err)
		}
	}()

	rl, err := ratelim.New(1000)
	if err != nil {
		return 0, fmt.Errorf("creating rate limiter: %s", err)
	}

	var lock sync.Mutex
	var modifiedCount int
	for r := range res.Next() {
		if r.Error != nil {
			return 0, fmt.Errorf("getting query record: %s", err)
		}

		// Key format: /ffs/manager/api/<api-id>/istore/storageconfig/<cid>
		key := datastore.NewKey(r.Key)
		parts := key.Namespaces()
		if len(parts) < 6 {
			continue
		}
		if parts[4] != "istore" || parts[5] != "cidstorageconfig" {
			continue
		}

		r := r
		rl.Exec(func() error {
			var cfg ffs.StorageConfig
			if err := json.Unmarshal(r.Value, &cfg); err != nil {
				return fmt.Errorf("unmarshaling storage config: %s", err)
			}

			modified, err := t(&cfg)
			if err != nil {
				return fmt.Errorf("applying transform: %s", err)
			}
			if !modified {
				return nil
			}

			lock.Lock()
			modifiedCount++
			lock.Unlock()
			if dryrun {
				return nil
			}

			buf, err := json.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshaling storage config: %s", err)
			}
			if err := ds.Put(datastore.NewKey(r.Key), buf); err != nil {
				return fmt.Errorf("put in datastore: %s", err)
			}

			return nil
		})
	}

	errors := rl.Wait()
	if len(errors) > 0 {
		for err := range errors {
			log.Error(err)
		}
		return 0, fmt.Errorf("finished with %d errors", len(errors))
	}

	return modifiedCount, nil
}
