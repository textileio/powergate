package dataprep

import (
	"context"
	"fmt"
	"io"
	"os"

	commcid "github.com/filecoin-project/go-fil-commcid"
	commP "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core/coreunix"
	ipld "github.com/ipfs/go-ipld-format"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

// CommP calculates the piece cid and size from an io.Reader.
func CommP(r io.Reader) (cid.Cid, uint64, error) {
	cp := &commP.Calc{}
	_, err := io.Copy(cp, r)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("copying data to aggregator: %s", err)
	}

	rawCommP, pieceSize, err := cp.Digest()
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("calculating final digest: %s", err)
	}
	pieceCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("converting commP to cid: %s", err)
	}

	return pieceCid, pieceSize, nil
}

// Dagify creates a UnixFS DAG from the provided file.
func Dagify(ctx context.Context, dagService ipld.DAGService, path string, progressBytes chan<- int64) (cid.Cid, error) {
	fileAdder, err := coreunix.NewAdder(ctx, nil, nil, dagService)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating unixfs adder: %s", err)
	}
	fileAdder.Pin = false
	fileAdder.Progress = true
	events := make(chan interface{}, 10)
	fileAdder.Out = events

	f, err := os.Open(path)
	if err != nil {
		return cid.Undef, fmt.Errorf("opening path: %s", err)
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return cid.Undef, fmt.Errorf("getting stat of data: %s", err)
	}
	fs, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating serial file: %s", err)
	}
	defer func() { _ = fs.Close() }()

	var (
		dagifyErr error
		dataCid   cid.Cid
	)
	go func() {
		defer close(events)

		nd, err := fileAdder.AddAllAndPin(fs)
		if err != nil {
			dagifyErr = err
			return
		}
		dataCid = nd.Cid()
	}()
	currentName := ""
	var previousSize int64
	for event := range events {
		output, ok := event.(*coreiface.AddEvent)
		if !ok {
			return cid.Undef, fmt.Errorf("unknown event type")
		}
		if stat.IsDir() && output.Name == "" {
			continue
		}
		if currentName != output.Name {
			currentName = output.Name
			previousSize = 0
		}
		if output.Bytes > 0 && progressBytes != nil {
			progressBytes <- (-previousSize + output.Bytes)
		}
		previousSize = output.Bytes
	}
	if dagifyErr != nil {
		return cid.Undef, fmt.Errorf("creating dag for data: %s", dagifyErr)
	}

	return dataCid, nil
}
