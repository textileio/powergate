package coreipfs

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	logging "github.com/ipfs/go-log/v2"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/textileio/fil-tools/fpa"
)

var (
	log = logging.Logger("coreipfs")
)

type CoreIpfs struct {
	ipfs iface.CoreAPI
}

var _ fpa.HotLayer = (*CoreIpfs)(nil)

func New(ipfs iface.CoreAPI) *CoreIpfs {
	return &CoreIpfs{
		ipfs: ipfs,
	}
}

func (ci *CoreIpfs) Add(ctx context.Context, r io.Reader) (cid.Cid, error) {
	path, err := ci.ipfs.Unixfs().Add(ctx, ipfsfiles.NewReaderFile(r), options.Unixfs.Pin(false))
	if err != nil {
		return cid.Undef, err
	}
	return path.Cid(), nil
}

func (ci *CoreIpfs) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	n, err := ci.ipfs.Unixfs().Get(ctx, path.IpfsPath(c))
	if err != nil {
		return nil, err
	}
	file := ipfsfiles.ToFile(n)
	if file == nil {
		return nil, fmt.Errorf("node is a directory")
	}
	return file, nil
}

func (ci *CoreIpfs) Pin(ctx context.Context, c cid.Cid) (fpa.HotInfo, error) {
	log.Infof("pinning %s", c)
	var i fpa.HotInfo
	pth := path.IpfsPath(c)
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err := ci.ipfs.Pin().Add(ctx, pth, options.Pin.Recursive(true)); err != nil {
		return i, err
	}
	stat, err := ci.ipfs.Block().Stat(ctx, pth)
	if err != nil {
		return i, err
	}
	i.Size = stat.Size()
	i.Ipfs = fpa.IpfsHotInfo{
		Created: time.Now(),
	}
	log.Infof("pinned %s successfully", c)
	return i, nil
}
