package dataprep

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commP "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core/coreunix"
	ipld "github.com/ipfs/go-ipld-format"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

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

func Dagify(ctx context.Context, dagService ipld.DAGService, path string) (cid.Cid, error) {
	events := make(chan interface{}, 10)

	fileAdder, err := coreunix.NewAdder(ctx, nil, nil, dagService)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating unixfs adder: %s", err)
	}
	fileAdder.Pin = false
	fileAdder.Progress = true
	fileAdder.Out = events

	f, err := os.Open(path)
	if err != nil {
		return cid.Undef, fmt.Errorf("opening path: %s", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return cid.Undef, fmt.Errorf("getting stat of data: %s", err)
	}
	fs, err := files.NewSerialFile(path, false, stat)
	if err != nil {
		return cid.Undef, fmt.Errorf("creating serial file: %s", err)
	}
	defer fs.Close()

	dataSize := int(stat.Size())
	if stat.IsDir() {
		dataSize = 0
		err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				dataSize += int(info.Size())
			}
			return err
		})
		if err != nil {
			return cid.Undef, fmt.Errorf("walking path: %s", err)
		}
	}
	bar := pb.StartNew(dataSize)
	bar.Set(pb.Bytes, true)

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
		if output.Bytes > 0 {
			bar.Add64(-previousSize + output.Bytes)
		}
		previousSize = output.Bytes
	}
	bar.Finish()
	if dagifyErr != nil {
		return cid.Undef, fmt.Errorf("creating dag for data: %s", dagifyErr)
	}

	return dataCid, nil
}
