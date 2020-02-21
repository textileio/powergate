package fastapi

import (
	"context"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/interface-go-ipfs-core/path"
	ftypes "github.com/textileio/fil-tools/fpa/types"
)

func (i *Instance) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	ar := i.auditer.Start(ctx, i.ID().String())
	ar.Close()
	r, err := i.get(ctx, ar, c)
	if err != nil {
		ar.Errored(err)
		return nil, err
	}
	ar.Success()
	return r, nil
}

func (i *Instance) get(ctx context.Context, oa ftypes.OpAuditer, c cid.Cid) (io.Reader, error) {
	n, err := i.ipfs.Unixfs().Get(ctx, path.IpfsPath(c))
	if err != nil {
		return nil, err
	}
	file := ipfsfiles.ToFile(n)
	if file == nil {
		return nil, fmt.Errorf("node is a directory")
	}
	return file, nil
}
