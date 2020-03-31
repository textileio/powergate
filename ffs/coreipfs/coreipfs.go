package coreipfs

import (
	"bytes"
	"context"
	"fmt"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	logging "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("ffs-coreipfs")
)

// CoreIpfs is an implementation of HotStorage interface which saves data
// into a remote go-ipfs using the HTTP API.
type CoreIpfs struct {
	ipfs iface.CoreAPI
}

var _ ffs.HotStorage = (*CoreIpfs)(nil)

// New returns a new CoreIpfs instance
func New(ipfs iface.CoreAPI) *CoreIpfs {
	return &CoreIpfs{
		ipfs: ipfs,
	}
}

// Put saves a Block.
func (ci *CoreIpfs) Put(ctx context.Context, b blocks.Block) error {
	if _, err := ci.ipfs.Block().Put(ctx, bytes.NewReader(b.RawData())); err != nil {
		return fmt.Errorf("adding block to ipfs node: %s", err)
	}
	return nil
}

// Stats return status information for a Cid.
func (ci *CoreIpfs) IsStored(ctx context.Context, c cid.Cid) (bool, error) {
	ci.ipfs.Block()
	pins, err := ci.ipfs.Pin().Ls(ctx)
	if err != nil {
		return false, fmt.Errorf("getting pins from IPFS: %s", err)
	}
	for _, p := range pins {
		if p.Path().Cid() == c {
			return true, nil
		}
	}
	return false, nil
}

// Add adds an io.Reader data as file in the IPFS node.
func (ci *CoreIpfs) Add(ctx context.Context, r io.Reader) (cid.Cid, error) {
	path, err := ci.ipfs.Unixfs().Add(ctx, ipfsfiles.NewReaderFile(r), options.Unixfs.Pin(false))
	if err != nil {
		return cid.Undef, fmt.Errorf("adding data to ipfs: %s", err)
	}
	return path.Cid(), nil
}

// Get retrieves a cid from the IPFS node.
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

// Pin pins as cid in the IPFS node
func (ci *CoreIpfs) Pin(ctx context.Context, c cid.Cid) (int, error) {
	pth := path.IpfsPath(c)
	if err := ci.ipfs.Pin().Add(ctx, pth, options.Pin.Recursive(true)); err != nil {
		return 0, fmt.Errorf("pinning cid %s: %s", c, err)
	}
	stat, err := ci.ipfs.Block().Stat(ctx, pth)
	if err != nil {
		return 0, fmt.Errorf("getting stats of cid %s: %s", c, err)
	}
	return stat.Size(), nil
}
