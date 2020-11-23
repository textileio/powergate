package migration

import (
	"encoding/json"
	"fmt"

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

func pinstoreFilling(txn datastore.Txn, cidOwners map[cid.Cid][]ffs.APIID) error {
	// This migration should fill Pinstore. Pinstore keeps track of
	// which IIDs are pinning a Cid in hot-storage.

	// Step 1/2:
	// - Iterate over all cidOwners and make a list of
	//   IIDs that are pinning a Cid in hot-storage.
	cidsPinstore := map[cid.Cid][]ffs.APIID{}
	for c, iids := range cidOwners {
		for _, iid := range iids {
			sc, err := v0GetStorageConfig(txn, iid, c)
			if err != nil {
				return fmt.Errorf("getting storage config: %s", err)
			}
			if sc.Hot.Enabled {
				cidsPinstore[c] = append(cidsPinstore[c], iid)
			}
		}
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
		if err := txn.Put(k, buf); err != nil {
			return fmt.Errorf("put in datastore: %s", err)
		}
	}

	return nil
}
