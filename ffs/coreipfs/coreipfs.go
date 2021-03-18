package coreipfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	logging "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/coreipfs/internal/pinstore"
	txndstr "github.com/textileio/powergate/v2/txndstransform"
)

var (
	log = logging.Logger("ffs-coreipfs")

	// ErrUnpinnedCid indicates that the operation failed because
	// the provided cid is unpinned.
	ErrUnpinnedCid = errors.New("can't unpin an unpinned cid")
	// ErrReplaceFromNotPinned indicates that the source cid to be replaced
	// isn't pinned.
	ErrReplaceFromNotPinned = errors.New("c1 isn't pinned")
)

// CoreIpfs is an implementation of HotStorage interface which saves data
// into a remote go-ipfs using the HTTP API.
type CoreIpfs struct {
	ipfs iface.CoreAPI
	ps   *pinstore.Store

	lock sync.Mutex
}

var _ ffs.HotStorage = (*CoreIpfs)(nil)

// New returns a new CoreIpfs instance.
func New(ds datastore.TxnDatastore, ipfs iface.CoreAPI, l ffs.JobLogger) (*CoreIpfs, error) {
	ps, err := pinstore.New(txndstr.Wrap(ds, "pinstore"))
	if err != nil {
		return nil, fmt.Errorf("loading pinstore: %s", err)
	}
	ci := &CoreIpfs{
		ipfs: ipfs,
		ps:   ps,
	}
	return ci, nil
}

// Stage adds the data of io.Reader in the storage, and creates a stage-pin on the resulting cid.
func (ci *CoreIpfs) Stage(ctx context.Context, iid ffs.APIID, r io.Reader) (cid.Cid, error) {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	p, err := ci.ipfs.Unixfs().Add(ctx, ipfsfiles.NewReaderFile(r), options.Unixfs.Pin(true))
	if err != nil {
		return cid.Undef, fmt.Errorf("adding data to ipfs: %s", err)
	}

	if err := ci.ps.AddStaged(iid, p.Cid()); err != nil {
		return cid.Undef, fmt.Errorf("saving new pin in pinstore: %s", err)
	}

	return p.Cid(), nil
}

// StageCid pull the Cid data and stage-pin it.
func (ci *CoreIpfs) StageCid(ctx context.Context, iid ffs.APIID, c cid.Cid) error {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	if err := ci.ipfs.Pin().Add(ctx, path.IpfsPath(c), options.Pin.Recursive(true)); err != nil {
		return fmt.Errorf("adding data to ipfs: %s", err)
	}

	if err := ci.ps.AddStaged(iid, c); err != nil {
		return fmt.Errorf("saving new pin in pinstore: %s", err)
	}

	return nil
}

// Get retrieves a cid data from the IPFS node.
func (ci *CoreIpfs) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	n, err := ci.ipfs.Unixfs().Get(ctx, path.IpfsPath(c))
	if err != nil {
		return nil, fmt.Errorf("getting cid %s from ipfs: %s", c, err)
	}
	file := ipfsfiles.ToFile(n)
	if file == nil {
		return nil, fmt.Errorf("node is a directory")
	}
	return file, nil
}

// Pin a cid for an APIID. If the cid was already pinned by a stage from APIID,
// the Cid is considered fully-pinned and not a candidate to be unpinned by GCStaged().
func (ci *CoreIpfs) Pin(ctx context.Context, iid ffs.APIID, c cid.Cid) (int, error) {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	p := path.IpfsPath(c)

	// If some APIID already pinned this Cid in the underlying go-ipfs node, then
	// we don't need to call the Pin API, just count the reference from this APIID.
	if !ci.ps.IsPinned(c) {
		if err := ci.ipfs.Pin().Add(ctx, p, options.Pin.Recursive(true)); err != nil {
			return 0, fmt.Errorf("pinning cid %s: %s", c, err)
		}
	}
	s, err := ci.ipfs.Object().Stat(ctx, p)
	if err != nil {
		return 0, fmt.Errorf("getting stats of cid %s: %s", c, err)
	}

	// Count +1 reference to this Cid by APIID.
	if err := ci.ps.Add(iid, p.Cid()); err != nil {
		return 0, fmt.Errorf("saving new pin in pinstore: %s", err)
	}

	return s.CumulativeSize, nil
}

// Unpin unpins a Cid for an APIID. If the Cid isn't pinned, it returns ErrUnpinnedCid.
func (ci *CoreIpfs) Unpin(ctx context.Context, iid ffs.APIID, c cid.Cid) error {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	return ci.removeAndUnpinIfApplies(ctx, iid, c)
}

// Replace moves the pin from c1 to c2. If c2 was already pinned from a stage,
// it's considered fully-pinned and not GCable.
func (ci *CoreIpfs) Replace(ctx context.Context, iid ffs.APIID, c1 cid.Cid, c2 cid.Cid) (int, error) {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	p1 := path.IpfsPath(c1)
	p2 := path.IpfsPath(c2)

	c1refcount, _ := ci.ps.RefCount(c1)
	c2refcount, _ := ci.ps.RefCount(c2)

	if c1refcount == 0 {
		return 0, ErrReplaceFromNotPinned
	}

	// If c1 has a single reference, which must be from iid...
	if c1refcount == 1 {
		// If c2 isn't pinned, then we can move the pin so to unpin c1 and pin c2.
		if c2refcount == 0 {
			if err := ci.ipfs.Pin().Update(ctx, p1, p2); err != nil {
				return 0, fmt.Errorf("updating pin %s to %s: %s", c1, c2, err)
			}
		} else { // If c2 is pinned, then we need to unpin c1 (c2 is already pinned by other iid).
			if err := ci.ipfs.Pin().Rm(ctx, path.IpfsPath(c1), options.Pin.RmRecursive(true)); err != nil {
				return 0, fmt.Errorf("unpinning cid from ipfs node: %s", err)
			}
		}
	} else if c2refcount == 0 {
		// - c1 is pinned by another iid, so we can't unpin it.
		// - c2 isn't pinned by anyone, so we should pin it.
		if err := ci.ipfs.Pin().Add(ctx, p2, options.Pin.Recursive(true)); err != nil {
			return 0, fmt.Errorf("pinning cid %s: %s", c2, err)
		}
	}
	// If none of the if branches applied:
	// - c1 is pinned by another iid, so we can't unpin it.
	// - c2 is pinned by some other iid, so it's already pinned in the node, no need to do it.

	// In any case of above if, update the ref counts.
	if err := ci.ps.Remove(iid, c1); err != nil {
		return 0, fmt.Errorf("removing cid in pinstore: %s", err)
	}
	if err := ci.ps.Add(iid, c2); err != nil {
		return 0, fmt.Errorf("adding cid in pinstore: %s", err)
	}

	stat, err := ci.ipfs.Object().Stat(ctx, p2)
	if err != nil {
		return 0, fmt.Errorf("getting stats of cid %s: %s", c2, err)
	}

	return stat.CumulativeSize, nil
}

// IsPinned returns true if c is pinned by iid.
func (ci *CoreIpfs) IsPinned(ctx context.Context, iid ffs.APIID, c cid.Cid) (bool, error) {
	return ci.ps.IsPinnedBy(iid, c), nil
}

// GCStaged unpins Cids that are only pinned by Stage() calls and all pins satisfy the filters.
func (ci *CoreIpfs) GCStaged(ctx context.Context, exclude []cid.Cid, olderThan time.Time) ([]cid.Cid, error) {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	unpinLst, err := ci.getGCCandidates(exclude, olderThan)
	if err != nil {
		return nil, fmt.Errorf("getting gc cid candidates: %s", err)
	}

	for _, c := range unpinLst {
		if err := ci.unpinStaged(ctx, c); err != nil {
			return nil, fmt.Errorf("unpinning cid from ipfs node: %s", err)
		}
	}

	return unpinLst, nil
}

// PinnedCids return detailed information about pinned cids.
func (ci *CoreIpfs) PinnedCids(ctx context.Context) ([]ffs.PinnedCid, error) {
	ci.lock.Lock()
	defer ci.lock.Unlock()

	ps, err := ci.ps.GetAll()
	if err != nil {
		return nil, fmt.Errorf("getting pins from pinstore: %s", err)
	}

	res := make([]ffs.PinnedCid, len(ps))
	for i, pc := range ps {
		npc := ffs.PinnedCid{
			Cid:    pc.Cid,
			APIIDs: make([]ffs.APIIDPinnedCid, len(pc.Pins)),
		}
		for j, upc := range pc.Pins {
			npc.APIIDs[j] = ffs.APIIDPinnedCid{
				ID:        upc.APIID,
				Staged:    upc.Staged,
				CreatedAt: upc.CreatedAt,
			}
		}
		res[i] = npc
	}

	return res, nil
}

func (ci *CoreIpfs) getGCCandidates(exclude []cid.Cid, olderThan time.Time) ([]cid.Cid, error) {
	lst, err := ci.ps.GetAllOnlyStaged()
	if err != nil {
		return nil, fmt.Errorf("get staged pins: %s", err)
	}

	excludeMap := map[cid.Cid]struct{}{}
	for _, c := range exclude {
		excludeMap[c] = struct{}{}
	}

	var unpinList []cid.Cid
Loop:
	for _, stagedPin := range lst {
		// Double check that ref count is consistent.
		total, staged := ci.ps.RefCount(stagedPin.Cid)
		if total != staged {
			return nil, fmt.Errorf("GC candidates are inconsistent")
		}

		// Skip Cids that are excluded.
		if _, ok := excludeMap[stagedPin.Cid]; ok {
			log.Infof("skipping staged cid %s since it's in exclusion list", stagedPin.Cid)
			continue Loop
		}
		// A Cid is only safe to GC if all existing stage-pin are older than
		// specified parameter. If any iid stage-pined the Cid more recently than olderThan
		// we still have to wait a bit more to consider it for GC.
		for _, sp := range stagedPin.Pins {
			if sp.CreatedAt > olderThan.Unix() {
				continue Loop
			}
		}

		// The Cid only has staged-pins, and all iids that staged it aren't in exclusion list
		// plus are older than olderThan ==> Safe to GCed.
		unpinList = append(unpinList, stagedPin.Cid)
	}

	return unpinList, nil
}

func (ci *CoreIpfs) removeAndUnpinIfApplies(ctx context.Context, iid ffs.APIID, c cid.Cid) error {
	count, _ := ci.ps.RefCount(c)
	if count == 0 {
		return ErrUnpinnedCid
	}

	if count == 1 {
		// There aren't more pinnings for this Cid, let's unpin from IPFS.
		log.Infof("unpinning cid %s with ref count 0", c)
		if err := ci.ipfs.Pin().Rm(ctx, path.IpfsPath(c), options.Pin.RmRecursive(true)); err != nil {
			return fmt.Errorf("unpinning cid from ipfs node: %s", err)
		}
	}

	if err := ci.ps.Remove(iid, c); err != nil {
		return fmt.Errorf("removing cid from pinstore: %s", err)
	}

	return nil
}

func (ci *CoreIpfs) unpinStaged(ctx context.Context, c cid.Cid) error {
	count, stagedCount := ci.ps.RefCount(c)

	// Just in case, verify that the total number of pins are equal
	// to stage-pins. That is, nobody is pinning this Cid apart from Stage() calls.
	if count != stagedCount {
		return fmt.Errorf("cid %s hasn't only stage-pins, total %d staged %d", c, count, stagedCount)
	}

	if err := ci.ipfs.Pin().Rm(ctx, path.IpfsPath(c), options.Pin.RmRecursive(true)); err != nil {
		return fmt.Errorf("unpinning cid from ipfs node: %s", err)
	}

	if err := ci.ps.RemoveStaged(c); err != nil {
		return fmt.Errorf("removing all staged pins for %s: %s", c, err)
	}

	return nil
}
