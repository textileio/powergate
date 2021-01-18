package api

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/scheduler"
)

// StorageInfo returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns ErrNotFound.
func (i *API) StorageInfo(cid cid.Cid) (ffs.StorageInfo, error) {
	inf, err := i.sched.GetStorageInfo(i.cfg.ID, cid)
	if err == scheduler.ErrNotFound {
		return inf, ErrNotFound
	}
	if err != nil {
		return inf, fmt.Errorf("getting cid storage info: %s", err)
	}
	return inf, nil
}

// ListStorageInfo returns a list of information about all stored cids, filtered by cids if provided.
func (i *API) ListStorageInfo(cids ...cid.Cid) ([]ffs.StorageInfo, error) {
	return i.sched.ListStorageInfo([]ffs.APIID{i.cfg.ID}, cids)
}
