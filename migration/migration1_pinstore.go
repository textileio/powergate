package migration

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/textileio/powergate/ffs"
)

// The following two datastructures are
// what Pinstore in v1 expects to be saved
// in the datastore. We can't use pinstore.PinnedCid
// struct directly because is an `internal` package.
type v1PinstorePinnedCid struct {
	Cid  cid.Cid
	Pins []v1PinstorePin
}
type v1PinstorePin struct {
	APIID     ffs.APIID
	Staged    bool
	CreatedAt int64
}

func pinstoreFilling(ds datastoreReaderWriter, cidOwners map[cid.Cid][]ffs.APIID) error {
	// This migration should fill Pinstore. Pinstore keeps track of
	// which IIDs are pinning a Cid in hot-storage.

	// Step 1/2:
	// - Iterate over all cidOwners and make a list of
	//   IIDs that are pinning a Cid in hot-storage.
	var lock sync.Mutex
	var errors []string
	lim := make(chan struct{}, 1000)
	cidsPinstore := map[cid.Cid][]ffs.APIID{}
	for c, iids := range cidOwners {
		lim <- struct{}{}
		c := c
		iids := iids
		go func() {
			defer func() { <-lim }()
			for _, iid := range iids {
				sc, err := v0GetStorageConfig(ds, iid, c)
				if err != nil {
					lock.Lock()
					errors = append(errors, fmt.Sprintf("getting storage config: %s", err))
					lock.Unlock()
					return
				}
				if sc.Hot.Enabled {
					lock.Lock()
					cidsPinstore[c] = append(cidsPinstore[c], iid)
					lock.Unlock()
				}
			}
		}()
	}

	for i := 0; i < cap(lim); i++ {
		lim <- struct{}{}
	}

	// Step 2/2:
	// - Include this generated data in Pinstore.
	for c, iids := range cidsPinstore {
		if len(iids) == 0 {
			continue
		}
		r := v1PinstorePinnedCid{
			Cid:  c,
			Pins: make([]v1PinstorePin, len(iids)),
		}
		for i := range iids {
			r.Pins[i] = v1PinstorePin{
				APIID:     iids[i],
				CreatedAt: 0,
			}
		}
		k := datastore.NewKey("/ffs/coreipfs/pinstore/pins/" + r.Cid.String())
		buf, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("marshaling to datastore: %s", err)
		}
		if err := ds.Put(k, buf); err != nil {
			return fmt.Errorf("put in datastore: %s", err)
		}
	}

	return nil
}
